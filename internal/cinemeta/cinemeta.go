package cinemeta

import (
	"strconv"
	"strings"

	"github.com/bongnv/jackett-stremio/internal/model"
	"github.com/go-resty/resty/v2"
)

type CineMeta struct {
	client *resty.Client
}

type MovieInfoResponse struct {
	Meta MetaInfo `json:"meta"`
}

type MetaInfo struct {
	Name   string `json:"name"`
	Year   string `json:"year"`
	IMDBID string `json:"imdb_id"`
}

func New() *CineMeta {
	return &CineMeta{
		client: resty.New().SetBaseURL("https://v3-cinemeta.strem.io"),
	}
}

func (c *CineMeta) GetMovieById(id string) (*model.MetaInfo, error) {
	resp, err := c.client.R().SetResult(&MovieInfoResponse{}).Get("/meta/movie/" + id + ".json")
	if err != nil {
		return nil, err
	}

	result := resp.Result().(*MovieInfoResponse)
	year, _ := strconv.Atoi(result.Meta.Year)
	imdbID, err := strconv.Atoi(strings.TrimPrefix(result.Meta.IMDBID, "tt"))

	return &model.MetaInfo{
		Name:     result.Meta.Name,
		IMDBID:   uint(imdbID),
		FromYear: year,
		ToYear:   year,
	}, nil
}

func (c CineMeta) GetSeriesById(id string) (*model.MetaInfo, error) {
	resp, err := c.client.R().SetResult(&MovieInfoResponse{}).Get("/meta/series/" + id + ".json")
	if err != nil {
		return nil, err
	}

	result := resp.Result().(*MovieInfoResponse)
	tokens := strings.Split(result.Meta.Year, "–")
	fromYear := 0
	toYear := 0
	if len(tokens) > 1 {
		fromYear, _ = strconv.Atoi(tokens[0])
		toYear, _ = strconv.Atoi(tokens[1])
	} else if len(tokens) > 0 {
		fromYear, _ = strconv.Atoi(tokens[0])
		toYear = fromYear
	}
	imdbID, err := strconv.Atoi(strings.TrimPrefix(result.Meta.IMDBID, "tt"))

	return &model.MetaInfo{
		Name:     result.Meta.Name,
		IMDBID:   uint(imdbID),
		FromYear: fromYear,
		ToYear:   toYear,
	}, nil
}
