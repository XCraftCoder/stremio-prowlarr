package jackett

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	moviesCategory = "2000"
	tvCategory     = "5000"
)

type Jackett struct {
	client *resty.Client
	apiURL string
}

func New(apiURL string, apiKey string) *Jackett {
	client := resty.New().
		SetBaseURL(apiURL).
		SetQueryParam("apikey", apiKey).
		SetRedirectPolicy(NoRedirectForMagnet())

	return &Jackett{
		client: client,
		apiURL: apiURL,
	}
}

func (j *Jackett) GetAllIndexers() ([]Indexer, error) {
	result := &IndexersResponse{}
	_, err := j.client.
		R().
		SetQueryParam("t", "indexers").
		SetQueryParam("configured", "true").
		SetResult(result).
		Get("/api/v2.0/indexers/all/results/torznab/api")

	if err != nil {
		return nil, err
	}

	return result.Indexers, nil
}

func (j *Jackett) SearchMovieTorrents(indexer Indexer, name string) ([]Torrent, error) {
	result := &TorrentsResponse{}
	_, err := j.client.
		R().
		SetQueryParam("t", "movie").
		SetQueryParam("query", name).
		SetQueryParam("category", moviesCategory).
		SetResult(result).
		Get("api/v2.0/indexers/" + indexer.ID + "/results")

	if err != nil {
		return nil, err
	}

	for i := range result.Torrents {
		result.Torrents[i].Link = strings.Replace(result.Torrents[i].Link, "http://localhost:9117", j.apiURL, 1)
	}

	return result.Torrents, nil
}

func (j *Jackett) FetchMagnetURI(torrent Torrent) (string, error) {
	if torrent.MagnetUri != "" {
		return torrent.MagnetUri, nil
	}

	resp, err := j.client.R().Get(torrent.Link)
	if err != nil {
		return "", err
	}

	return resp.Header().Get("location"), nil
}
