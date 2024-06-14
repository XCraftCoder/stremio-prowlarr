package jackett

import (
	"bytes"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2/log"
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
		SetRedirectPolicy(NotFollowMagnet())

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
		result.Torrents[i].InfoHash = strings.ToLower(result.Torrents[i].InfoHash)
	}

	return result.Torrents, nil
}

func (j *Jackett) FetchMagnetURI(torrent Torrent) (Torrent, error) {
	if torrent.MagnetUri == "" {
		resp, err := j.client.R().Get(torrent.Link)
		if err != nil {
			log.Errorf("Failed to fetch magnet link for %s due to: %v", torrent.Link, err)
			return torrent, err
		}

		if resp.Header().Get("Content-Type") == "application/x-bittorrent" {
			torFile, err := parseTorrentFile(bytes.NewReader(resp.Body()))
			if err != nil {
				log.Errorf("Invalid torrent file for %s with: %v", torrent.Link, err)
				return torrent, err
			}

			magnet := &Magnet{
				Name:     torrent.Title,
				InfoHash: torFile.Info.Hash,
				Trackers: torFile.AnnounceList,
			}
			torrent.MagnetUri = magnet.String()
			torrent.InfoHash = strings.ToLower(magnet.InfoHashStr())
		} else {
			torrent.MagnetUri = resp.Header().Get("location")
		}

		if torrent.MagnetUri == "" {
			log.Errorf("Unexpected magnet uri for %s", torrent.Guid)
			return torrent, errors.New("magnet uri is expected but not found")
		}
	}

	if torrent.InfoHash == "" {
		magnet, err := ParseMagnetUri(torrent.MagnetUri)
		if err != nil {
			return torrent, err
		}
		torrent.InfoHash = strings.ToLower(magnet.InfoHashStr())
	}

	return torrent, nil
}
