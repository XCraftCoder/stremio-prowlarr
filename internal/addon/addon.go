package addon

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/bongnv/jackett-stremio/internal/cinemeta"
	"github.com/bongnv/jackett-stremio/internal/debrid/realdebrid"
	"github.com/bongnv/jackett-stremio/internal/jackett"
	"github.com/bongnv/jackett-stremio/internal/model"
	"github.com/bongnv/jackett-stremio/internal/pipe"
	"github.com/bongnv/jackett-stremio/internal/titleparser"
	"github.com/coocood/freecache"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	cacheSize          = 50 * 1024 * 1024 // 50MB
	streamRecordExpiry = 10 * 60          // 10m
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
)

// Addon implements a Stremio addon
type Addon struct {
	id          string
	name        string
	version     string
	description string

	cinemetaClient *cinemeta.CineMeta
	jackettClient  *jackett.Jackett
	realDebrid     *realdebrid.RealDebrid
	cache          *freecache.Cache
}

type Option func(*Addon)

type GetStreamsResponse struct {
	Streams []StreamItem `json:"streams"`
}

type streamRecord struct {
	ContentType   ContentType
	ID            string
	Season        int
	Episode       int
	HostURL       string
	RemoteAddress string
	MetaInfo      *model.MetaInfo
	TitleInfo     *titleparser.MetaInfo
	Indexer       jackett.Indexer
	Torrent       jackett.Torrent
	Files         []realdebrid.File
}

const (
	maxStreamsResult = 10
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

	if addon.jackettClient == nil {
		panic("jackett client must be provided")
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
	infoHash := strings.ToLower(c.Params("infoHash"))
	download, err := add.realDebrid.GetDownloadByInfoHash(infoHash)

	if err != nil && err != realdebrid.ErrNoTorrentFound {
		log.Errorf("Couldn't find download link from the given hash %s: %v", infoHash, err)
		return err
	}

	if err == realdebrid.ErrNoTorrentFound {
		val, err := add.cache.Get([]byte(infoHash))
		if err != nil {
			log.Errorf("Couldn't find the infoHash from cache: %v", err)
			return err
		}

		record := streamRecord{}
		err = gob.NewDecoder(bytes.NewReader(val)).Decode(&record)
		if err != nil {
			log.WithContext(c.Context()).Errorf("Couldn't decode the cached data: %v", err)
			return err
		}

		download, err = add.realDebrid.GetDownloadByMagnetURI(record.Torrent.MagnetUri)
		if err != nil {
			log.WithContext(c.Context()).Errorf("Couldn't add magnet link: %v", err)
			return err
		}
	}

	return c.Redirect(download)
}

func (add *Addon) HandleGetStreams(c *fiber.Ctx) error {
	p := pipe.New(add.sourceFromContext(c))

	p.Map(add.fetchMetaInfo)
	p.FanOut(add.fanOutToAllIndexers)
	p.FanOut(add.searchForTorrents)
	p.Map(add.parseTorrentTitle)
	p.Filter(excludeTorrents)
	p.Shuffle(hasMoreSeeders)
	p.FanOut(add.enrichInfoHash)
	p.Batch(add.filterByCached)
	p.Shuffle(hasHigherQuality)

	records := add.sinkResults(p)
	results := make([]StreamItem, 0, len(records))
	for _, r := range records {
		results = append(results, StreamItem{
			Name:  fmt.Sprintf("[%dp]", r.TitleInfo.Resolution),
			Title: fmt.Sprintf("%s\n%d|%d|%s", r.Torrent.Title, r.Torrent.Size, r.Torrent.Seeders, r.Indexer.Title),
			URL:   r.HostURL + "/download/" + r.Torrent.InfoHash,
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
			HostURL:       c.Protocol() + "://" + c.Hostname() + c.Path(),
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
	allIndexers, err := add.jackettClient.GetAllIndexers()
	if err != nil {
		return nil, fmt.Errorf("couldn't load all indexers: %v", err)
	}

	records := make([]*streamRecord, 0, len(allIndexers))
	for _, indexer := range allIndexers {
		newR := *r
		newR.Indexer = indexer
		records = append(records, &newR)
	}

	return records, nil
}

func (add *Addon) searchForTorrents(r *streamRecord) ([]*streamRecord, error) {
	torrents := []jackett.Torrent{}
	var err error
	switch r.ContentType {
	case ContentTypeMovie:
		torrents, err = add.jackettClient.SearchMovieTorrents(r.Indexer, r.MetaInfo.Name)
	}

	if err != nil {
		log.Errorf("Failed to load torrents for %s in %s, due to: %v", r.MetaInfo.Name, r.Indexer.ID, err)
		return nil, nil
	}

	records := make([]*streamRecord, 0, len(torrents))
	for _, torrent := range torrents {
		newRecord := *r
		newRecord.Torrent = torrent
		records = append(records, &newRecord)
	}

	log.Infof("Found %d from %s", len(records), r.Indexer.ID)
	return records, nil
}

func (add *Addon) enrichInfoHash(r *streamRecord) ([]*streamRecord, error) {
	var err error
	r.Torrent, err = add.jackettClient.FetchMagnetURI(r.Torrent)
	if err != nil {
		log.Errorf("Failed to fetch magnetUri for %s due to: %v", r.Torrent.Guid, err)
		return nil, nil
	}

	return []*streamRecord{r}, nil
}

func (add *Addon) filterByCached(records []*streamRecord) ([]*streamRecord, error) {
	infoHashs := make([]string, 0, len(records))
	for _, record := range records {
		if record.Torrent.InfoHash == "" {
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

		err = add.cache.Set([]byte(r.Torrent.InfoHash), buf.Bytes(), 10*60)
		if err != nil {
			log.Errorf("Failed to cache the record: %s, %v", r.Torrent.InfoHash, err)
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

func excludeTorrents(r *streamRecord) bool {
	return !slices.Contains(remuxSources, r.TitleInfo.Source) && !slices.Contains(camSources, r.TitleInfo.Source) && !r.TitleInfo.ThreeD
}

func hasMoreSeeders(r1, r2 *streamRecord) bool {
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
