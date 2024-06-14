package addon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bongnv/jackett-stremio/internal/cinemeta"
	"github.com/bongnv/jackett-stremio/internal/jackett"
	"github.com/bongnv/jackett-stremio/internal/model"
	"github.com/bongnv/jackett-stremio/internal/pipe"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Addon implements a Stremio addon
type Addon struct {
	id          string
	name        string
	version     string
	description string

	cinemetaClient *cinemeta.CineMeta
	jackettClient  *jackett.Jackett
}

type Option func(*Addon)

type GetStreamsResponse struct {
	Streams []StreamItem `json:"streams"`
}

type streamRecord struct {
	contentType   ContentType
	id            string
	season        int
	episode       int
	hostURL       string
	remoteAddress string
	metaInfo      *model.MetaInfo
	indexer       jackett.Indexer
	torrent       jackett.Torrent
}

const (
	maxStreamsResult = 10
)

func New(opts ...Option) *Addon {
	addon := &Addon{
		version:        "0.1.0",
		description:    "A Stremio addon",
		cinemetaClient: cinemeta.New(),
	}

	for _, opt := range opts {
		opt(addon)
	}

	if addon.jackettClient == nil {
		panic("jackett client must be provided")
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

func (add *Addon) HandleGetStreams(c *fiber.Ctx) error {
	p := pipe.New(add.sourceFromContext(c))

	p.Map(add.fetchMetaInfo)
	p.FanOut(add.fanOutToAllIndexers)
	p.FanOut(add.searchForTorrents)

	results := make([]StreamItem, 0, maxStreamsResult)
	err := p.Sink(func(r streamRecord) error {
		if len(results) == maxStreamsResult {
			log.Info("Enough results have been collected.")
			return nil
		}

		results = append(results, StreamItem{
			Name:  "Movie",
			Title: fmt.Sprintf("%s|%d|%d|%s", r.torrent.Title, r.torrent.Size, r.torrent.Seeders, r.indexer.Title),
			URL:   r.hostURL + "/download",
		})

		if len(results) == maxStreamsResult {
			p.Stop()
		}
		return nil
	})

	if err != nil {
		log.Errorf("Error while processing: %v", err)
	}

	return c.JSON(GetStreamsResponse{
		Streams: results,
	})
}

func (add *Addon) sourceFromContext(c *fiber.Ctx) func() ([]streamRecord, error) {
	return func() ([]streamRecord, error) {
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

		return []streamRecord{{
			contentType:   contentType,
			id:            id,
			season:        season,
			episode:       episode,
			hostURL:       c.Protocol() + "://" + c.Hostname() + "/" + c.Path(),
			remoteAddress: c.Context().RemoteIP().String(),
		}}, nil
	}
}

func (add *Addon) fetchMetaInfo(r streamRecord) (streamRecord, error) {
	switch r.contentType {
	case ContentTypeMovie:
		resp, err := add.cinemetaClient.GetMovieById(r.id)
		if err != nil {
			return r, err
		}

		r.metaInfo = resp
		return r, nil
	case ContentTypeSeries:
		resp, err := add.cinemetaClient.GetSeriesById(r.id)
		if err != nil {
			return r, err
		}

		r.metaInfo = resp
		return r, nil
	default:
		return r, errors.New("not supported content type")
	}
}

func (add *Addon) fanOutToAllIndexers(r streamRecord) ([]streamRecord, error) {
	allIndexers, err := add.jackettClient.GetAllIndexers()
	if err != nil {
		return nil, fmt.Errorf("couldn't load all indexers: %v", err)
	}

	records := make([]streamRecord, 0, len(allIndexers))
	for _, indexer := range allIndexers {
		r := r
		r.indexer = indexer
		records = append(records, r)
	}

	return records, nil
}

func (add *Addon) searchForTorrents(r streamRecord) ([]streamRecord, error) {
	torrents := []jackett.Torrent{}
	var err error
	switch r.contentType {
	case ContentTypeMovie:
		torrents, err = add.jackettClient.SearchMovieTorrents(r.indexer, r.metaInfo.Name)
	}

	if err != nil {
		log.Errorf("Failed to load torrents for %s in %s", r.metaInfo.Name, r.indexer.Title)
		return []streamRecord{}, nil
	}

	records := make([]streamRecord, 0, len(torrents))
	for _, torrent := range torrents {
		newRecord := r
		newRecord.torrent = torrent
		records = append(records, newRecord)
	}
	return records, nil
}
