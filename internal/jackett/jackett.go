package jackett

import (
	"github.com/go-resty/resty/v2"
)

const (
	moviesCategory = "2000"
	tvCategory     = "5000"
)

type Jackett struct {
	client *resty.Client
}

func New(apiURL string, apiKey string) *Jackett {
	client := resty.New().
		SetBaseURL(apiURL).
		SetQueryParam("apikey", apiKey)

	return &Jackett{
		client: client,
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

	return result.Torrents, nil
}
