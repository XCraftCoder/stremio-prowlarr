package jackett

import (
	"log"

	"github.com/go-resty/resty/v2"
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
	resp, err := j.client.
		R().
		SetQueryParam("t", "indexers").
		SetQueryParam("configured", "true").
		SetResult(result).
		Get("/api/v2.0/indexers/all/results/torznab/api")

	if err != nil {
		return nil, err
	}

	log.Println("resp: ", resp.String(), result.Indexers)
	return result.Indexers, nil
}
