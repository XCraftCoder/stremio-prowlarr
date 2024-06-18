package addon

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/bongnv/prowlarr-stremio/internal/cinemeta"
	"github.com/bongnv/prowlarr-stremio/internal/debrid/realdebrid"
	"github.com/bongnv/prowlarr-stremio/internal/model"
	"github.com/bongnv/prowlarr-stremio/internal/pipe"
	"github.com/bongnv/prowlarr-stremio/internal/prowlarr"
	"github.com/bongnv/prowlarr-stremio/internal/titleparser"
	"github.com/coocood/freecache"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	cacheSize          = 50 * 1024 * 1024 // 50MB
	streamRecordExpiry = 10 * 60          // 10m
	minTitleMatch      = 0.5
)

var (
	remuxSources = []string{
		"bdremux",
		"brremux",
		"webremux",
		"dlremux",
	}

	camSources = []string{
		"telesync",
		"cam",
		"hdcam",
	}

	mediaContainerExtensions = []string{
		"mkv",
		"mk3d",
		"mp4",
		"m4v",
		"mov",
		"avi",
	}

	nonWordCharacter = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// Addon implements a Stremio addon
type Addon struct {
	id          string
	name        string
	version     string
	description string

	cinemetaClient *cinemeta.CineMeta
	prowlarrClient *prowlarr.Prowlarr
	realDebrid     *realdebrid.RealDebrid
	cache          *freecache.Cache
}

type Option func(*Addon)

type GetStreamsResponse struct {
	Streams []StreamItem `json:"streams"`
}

type streamRecord struct {
	ContentType    ContentType
	ID             string
	Season         int
	Episode        int
	BaseURL        string
	RemoteAddress  string
	MetaInfo       *model.MetaInfo
	TitleInfo      *titleparser.MetaInfo
	Indexer        *prowlarr.Indexer
	Torrent        *prowlarr.Torrent
	Files          []*realdebrid.File
	MediaFile      *realdebrid.File
	SearchBySeason bool
}

const (
	maxStreamsResult = 10
	magnetUriExpiry  = 10 * 60 // 10 minutes
)

func New(opts ...Option) *Addon {
	addon := &Addon{
		version:        "0.1.0",
		description:    "A Stremio addon",
		cinemetaClient: cinemeta.New(),
		cache:          freecache.NewCache(cacheSize),
	}

	for _, opt := range opts {
		opt(addon)
	}

	if addon.prowlarrClient == nil {
		panic("prowlarr client must be provided")
	}

	if addon.realDebrid == nil {
		panic("realdebrid client must be provided")
	}

	return addon
}

func (add *Addon) HandleGetManifest(c *fiber.Ctx) error {
	manifest := &Manifest{
		ID:          add.id,
		Name:        add.name,
		Description: add.description,
		Version:     add.version,
		ResourceItems: []ResourceItem{
			{
				Name:       ResourceStream,
				Types:      []ContentType{ContentTypeMovie, ContentTypeSeries},
				IDPrefixes: []string{"tt"},
			},
		},
		Types:      []ContentType{ContentTypeMovie, ContentTypeSeries},
		Catalogs:   []CatalogItem{},
		IDPrefixes: []string{"tt"},
	}

	return c.JSON(manifest)
}

func (add *Addon) HandleDownload(c *fiber.Ctx) error {
	torrentID := strings.ToLower(c.Params("torrentID"))
	fileID := strings.ToLower(c.Params("fileID"))
	rawTorrentID, err := prowlarr.TorrentIDFromString(torrentID)
	if err != nil {
		log.Errorf("Invalid torrent ID %s, err: %v", torrentID, err)
		return err
	}

	magnetURI, err := add.cache.Get(rawTorrentID)
	if err != nil {
		log.WithContext(c.Context()).Errorf("Couldn't find the magnet URI for %s in cache: %v", torrentID, err)
		return err
	}

	magnet, err := prowlarr.ParseMagnetUri(string(magnetURI))
	if err != nil {
		log.WithContext(c.Context()).Errorf("Couldn't parse the magnet URI for %s: %v", torrentID, err)
		return err
	}

	download, err := add.realDebrid.GetDownloadByMagnetURI(magnet.InfoHashStr(), string(magnetURI), fileID)
	if err != nil {
		log.WithContext(c.Context()).Errorf("Couldn't generate the download link for %s, %s: %v", torrentID, fileID, err)
		return err
	}

	return c.Redirect(download)
}

func (add *Addon) HandleGetStreams(c *fiber.Ctx) error {
	compiled := regexp.MustCompile(`/stream/(movie|series).+$`)
	p := pipe.New(add.sourceFromContext(c))

	p.Map(add.fetchMetaInfo)
	p.FanOut(add.fanOutToAllIndexers)
	p.FanOut(includeSearchBySeason)
	p.FanOut(add.searchForTorrents)
	p.Map(add.parseTorrentTitle)
	p.Filter(excludeTorrents)
	p.Shuffle(hasMoreSeeders)
	p.FanOut(add.enrichInfoHash)
	p.Filter(deduplicateTorrent())
	p.Batch(add.enrichWithCachedFiles)
	p.FanOut(add.locateMediaFile)
	p.Shuffle(hasHigherQuality)

	records := add.sinkResults(p)
	results := make([]StreamItem, 0, len(records))
	for _, r := range records {
		results = append(results, StreamItem{
			Name:  fmt.Sprintf("[%dp]", r.TitleInfo.Resolution),
			Title: fmt.Sprintf("%s\n%s\n%s|%d|%s", r.Torrent.Title, r.MediaFile.FileName, bytesConvert(r.MediaFile.FileSize), r.Torrent.Seeders, r.Indexer.Name),
			URL:   r.BaseURL + compiled.ReplaceAllString(c.Path(), "/download/"+r.Torrent.GID.ToString()+"/"+r.MediaFile.ID),
		})
	}

	return c.JSON(GetStreamsResponse{
		Streams: results,
	})
}

func (add *Addon) sourceFromContext(c *fiber.Ctx) func() ([]*streamRecord, error) {
	return func() ([]*streamRecord, error) {
		id := c.Params("id")
		season := 0
		episode := 0
		contentType := ContentType(c.Params("type"))
		if contentType == ContentTypeSeries {
			tokens := strings.Split(id, "%3A")
			if len(tokens) != 3 {
				return nil, errors.New("invalid stream id")
			}
			id = tokens[0]
			season, _ = strconv.Atoi(tokens[1])
			episode, _ = strconv.Atoi(tokens[2])
		}

		return []*streamRecord{{
			ContentType:   contentType,
			ID:            id,
			Season:        season,
			Episode:       episode,
			BaseURL:       c.BaseURL(),
			RemoteAddress: c.Context().RemoteIP().String(),
		}}, nil
	}
}

func (add *Addon) fetchMetaInfo(r *streamRecord) (*streamRecord, error) {
	switch r.ContentType {
	case ContentTypeMovie:
		resp, err := add.cinemetaClient.GetMovieById(r.ID)
		if err != nil {
			return r, err
		}

		r.MetaInfo = resp
		return r, nil
	case ContentTypeSeries:
		resp, err := add.cinemetaClient.GetSeriesById(r.ID)
		if err != nil {
			return r, err
		}

		r.MetaInfo = resp
		return r, nil
	default:
		return r, errors.New("not supported content type")
	}
}

func (add *Addon) fanOutToAllIndexers(r *streamRecord) ([]*streamRecord, error) {
	allIndexers, err := add.prowlarrClient.GetAllIndexers()
	if err != nil {
		return nil, fmt.Errorf("couldn't load all indexers: %v", err)
	}

	records := make([]*streamRecord, 0, len(allIndexers))
	for _, indexer := range allIndexers {
		if !indexer.Enable {
			log.Infof("Skip %s as it's disabled", indexer.Name)
			continue
		}

		newR := *r
		newR.Indexer = indexer
		records = append(records, &newR)
	}

	return records, nil
}

func (add *Addon) searchForTorrents(r *streamRecord) ([]*streamRecord, error) {
	torrents := []*prowlarr.Torrent{}
	var err error

	switch r.ContentType {
	case ContentTypeMovie:
		torrents, err = add.prowlarrClient.SearchMovieTorrents(r.Indexer, r.MetaInfo.Name)
	case ContentTypeSeries:
		if r.SearchBySeason {
			torrents, err = add.prowlarrClient.SearchSeasonTorrents(r.Indexer, r.MetaInfo.Name, r.Season)
		} else {
			torrents, err = add.prowlarrClient.SearchSeriesTorrents(r.Indexer, r.MetaInfo.Name)
		}
	}

	if err != nil {
		log.Errorf("Failed to load torrents for %s in %s, due to: %v", r.MetaInfo.Name, r.Indexer.Name, err)
		return nil, nil
	}

	records := make([]*streamRecord, 0, len(torrents))
	for _, torrent := range torrents {
		newRecord := *r
		newRecord.Torrent = torrent
		records = append(records, &newRecord)
	}

	log.Infof("Found %d from %s", len(records), r.Indexer.Name)
	return records, nil
}

func (add *Addon) enrichInfoHash(r *streamRecord) ([]*streamRecord, error) {
	var err error

	if r.Torrent.MagnetUri == "" {
		magnetUri, err := add.cache.Get(r.Torrent.GID)
		if err == nil {
			r.Torrent.MagnetUri = string(magnetUri)
		}
	}

	r.Torrent, err = add.prowlarrClient.FetchMagnetURI(r.Torrent)
	if err != nil {
		log.Errorf("Failed to fetch magnetUri for %s due to: %v", r.Torrent.Guid, err)
		return nil, nil
	}

	if r.Torrent.MagnetUri == "" {
		log.Warnf("Unable to find Magnet URI for %s", r.Torrent.Guid)
		return nil, nil
	}

	err = add.cache.Set(r.Torrent.GID, []byte(r.Torrent.MagnetUri), magnetUriExpiry)
	if err != nil {
		log.Errorf("Failed to cache the magnet URI due to: %v", err)
		return nil, nil
	}

	return []*streamRecord{r}, nil
}

func (add *Addon) enrichWithCachedFiles(records []*streamRecord) ([]*streamRecord, error) {
	infoHashs := make([]string, 0, len(records))
	for _, record := range records {
		if record.Torrent.InfoHash == "" {
			log.Infof("Skipped %s due to missing infoHash", record.Torrent.Title)
			continue
		}

		infoHashs = append(infoHashs, record.Torrent.InfoHash)
	}

	filesByHash, err := add.realDebrid.GetFiles(infoHashs)
	if err != nil {
		log.Errorf("Failed to fetch files from debrid: %v", err)
		return nil, nil
	}

	cachedRecords := make([]*streamRecord, 0, len(records))
	for _, r := range records {
		if files, ok := filesByHash[r.Torrent.InfoHash]; ok {
			newR := *r
			newR.Files = files
			cachedRecords = append(cachedRecords, &newR)
			// } else {
			// 	log.Infof("Skipped %s due to missing cached file", r.Torrent.Title)
		}
	}

	log.Infof("Found %d cached from %d records", len(cachedRecords), len(records))
	return cachedRecords, nil
}

func (add *Addon) sinkResults(p *pipe.Pipe[streamRecord]) []*streamRecord {
	records := make([]*streamRecord, 0, maxStreamsResult)
	err := p.Sink(func(r *streamRecord) error {
		if len(records) == maxStreamsResult {
			log.Info("Enough results have been collected.")
			return nil
		}

		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(r)
		if err != nil {
			log.Errorf("Failed to encode: %v", err)
			return nil
		}

		records = append(records, r)
		if len(records) == maxStreamsResult {
			p.Stop()
		}
		return nil
	})

	if err != nil {
		log.Errorf("Error while processing: %v", err)
	}

	slices.SortFunc(records, cmpLowerQuality)
	return records
}

func (add *Addon) parseTorrentTitle(r *streamRecord) (*streamRecord, error) {
	r.TitleInfo = titleparser.Parse(r.Torrent.Title)
	return r, nil
}

func (add *Addon) locateMediaFile(r *streamRecord) ([]*streamRecord, error) {
	switch r.ContentType {
	case ContentTypeMovie:
		r.MediaFile = findMovieMediaFile(r.Files)
	case ContentTypeSeries:
		r.MediaFile = findEpisodeMediaFile(r.Files, fmt.Sprintf(`S%02dE%02d`, r.Season, r.Episode))

		if r.MediaFile == nil {
			r.MediaFile = findEpisodeMediaFile(r.Files, fmt.Sprintf(`\b%dx%02d\b`, r.Season, r.Episode))
		}

		if r.MediaFile == nil {
			r.MediaFile = findEpisodeMediaFile(r.Files, fmt.Sprintf(`(?i)\bS?%02d\b.+\bE?%02d\b`, r.Season, r.Episode))
		}

		if r.MediaFile == nil {
			r.MediaFile = findEpisodeMediaFile(r.Files, fmt.Sprintf(`\b%d\b`, r.Episode))
		}
	}

	if r.MediaFile != nil {
		return []*streamRecord{r}, nil
	}

	log.Infof("Couldn't locate media file: %s", r.Torrent.Title)
	// for _, f := range r.Files {
	// 	log.Infof("File %s", f.FileName)
	// }
	return nil, nil
}

func includeSearchBySeason(r *streamRecord) ([]*streamRecord, error) {
	if r.ContentType == ContentTypeSeries {
		newR := *r
		newR.SearchBySeason = true
		return []*streamRecord{
			&newR,
			r,
		}, nil
	}

	return []*streamRecord{r}, nil
}

func deduplicateTorrent() func(r *streamRecord) bool {
	found := &sync.Map{}
	return func(r *streamRecord) bool {
		if r.Torrent.InfoHash == "" {
			log.Infof("Skipped %s due to empty hash", r.Torrent.Title)
			return false
		}

		if _, loaded := found.LoadOrStore(r.Torrent.InfoHash, struct{}{}); loaded {
			log.Infof("Skipped %s due to duplication of %s", r.Torrent.Title, r.Torrent.InfoHash)
			return false
		}

		return true
	}
}

func findEpisodeMediaFile(files []*realdebrid.File, pattern string) *realdebrid.File {
	var mediaFile *realdebrid.File
	compiled := regexp.MustCompile(pattern)
	for _, f := range files {
		if !hasMediaExtension(f.FileName) || !compiled.MatchString(f.FileName) {
			continue
		}

		if mediaFile == nil || mediaFile.FileSize < f.FileSize {
			mediaFile = f
		}
	}

	return mediaFile
}

func findMovieMediaFile(files []*realdebrid.File) *realdebrid.File {
	var mediaFile *realdebrid.File
	for _, f := range files {
		if !hasMediaExtension(f.FileName) {
			continue
		}

		if mediaFile == nil || mediaFile.FileSize < f.FileSize {
			mediaFile = f
		}
	}

	return mediaFile
}

func hasMediaExtension(fileName string) bool {
	fileName = strings.ToLower(fileName)
	for _, extension := range mediaContainerExtensions {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}

	return false
}

func excludeTorrents(r *streamRecord) bool {
	qualityOK := !slices.Contains(remuxSources, r.TitleInfo.Quality) &&
		!slices.Contains(camSources, r.TitleInfo.Quality) && !r.TitleInfo.ThreeD
	imdbOK := (r.Torrent.Imdb == 0 || r.Torrent.Imdb == r.MetaInfo.IMDBID)
	titleOK := (r.Torrent.Imdb > 0 || checkTitleSimilarity(r.MetaInfo.Name, r.TitleInfo.Title) > minTitleMatch)
	yearOK := (r.TitleInfo.Year == 0 || (r.MetaInfo.FromYear <= r.TitleInfo.Year && r.MetaInfo.ToYear >= r.TitleInfo.Year))
	seasonOK := r.ContentType != ContentTypeSeries || (r.TitleInfo.FromSeason == 0 || (r.TitleInfo.FromSeason <= r.Season && r.TitleInfo.ToSeason >= r.Season))
	episodeOK := r.ContentType != ContentTypeSeries || (r.TitleInfo.Episode == 0 || r.TitleInfo.Episode == r.Episode)
	result := qualityOK && imdbOK && yearOK && seasonOK && episodeOK && titleOK
	// if !result {
	// 	log.Infof("Excluded %s, quality: %v, imdb: %v, year: %v, season: %v, episode: %v, title: %v",
	// 		r.Torrent.Title,
	// 		qualityOK,
	// 		imdbOK, yearOK,
	// 		seasonOK,
	// 		episodeOK,
	// 		titleOK,
	// 	)
	// }
	return result
}

func checkTitleSimilarity(left, right string) float64 {
	left = nonWordCharacter.ReplaceAllString(strings.ToLower(left), " ")
	right = nonWordCharacter.ReplaceAllString(strings.ToLower(right), " ")
	return strutil.Similarity(left, right, metrics.NewLevenshtein())
}

func hasMoreSeeders(r1, r2 *streamRecord) bool {
	if r1.Torrent.Imdb > 0 && r2.Torrent.Imdb == 0 {
		return true
	}

	if r1.Torrent.Imdb == 0 && r2.Torrent.Imdb > 0 {
		return false
	}

	if r1.TitleInfo.Resolution > r2.TitleInfo.Resolution {
		return true
	}

	if r1.TitleInfo.Resolution < r2.TitleInfo.Resolution {
		return false
	}

	if r1.Torrent.MagnetUri != "" && r2.Torrent.MagnetUri == "" {
		return true
	}

	if r1.Torrent.MagnetUri == "" && r2.Torrent.MagnetUri != "" {
		return false
	}

	return r1.Torrent.Seeders > r2.Torrent.Seeders
}

func hasHigherQuality(r1, r2 *streamRecord) bool {
	return cmpLowerQuality(r1, r2) != 1
}

func cmpLowerQuality(r1, r2 *streamRecord) int {
	if r1.TitleInfo.Resolution > r2.TitleInfo.Resolution {
		return -1
	}

	if r1.TitleInfo.Resolution < r2.TitleInfo.Resolution {
		return 1
	}

	if r1.Torrent.Size > r2.Torrent.Size {
		return -1
	}

	if r1.Torrent.Size < r2.Torrent.Size {
		return 1
	}

	return 0
}
