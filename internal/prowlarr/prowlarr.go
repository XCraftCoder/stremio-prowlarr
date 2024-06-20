package prowlarr

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	moviesCategory = "2000"
	tvCategory     = "5000"
)

type Prowlarr struct {
	client *resty.Client
	apiURL string
}

func New(apiURL string, apiKey string) *Prowlarr {
	client := resty.New().
		// SetDebug(true).
		SetBaseURL(apiURL).
		SetHeader("X-Api-Key", apiKey).
		SetRedirectPolicy(NotFollowMagnet())

	return &Prowlarr{
		client: client,
		apiURL: apiURL,
	}
}

func (j *Prowlarr) GetAllIndexers() ([]*Indexer, error) {
	result := []*Indexer{}
	resp, err := j.client.
		R().
		SetResult(&result).
		Get("/api/v1/indexer")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error response from prowlarr: %v", resp.Error())
	}

	return result, nil
}

func (j *Prowlarr) SearchMovieTorrents(indexer *Indexer, name string) ([]*Torrent, error) {
	result := []*Torrent{}
	resp, err := j.client.
		R().
		SetQueryParam("query", name).
		SetQueryParam("categories", moviesCategory).
		SetQueryParam("type", "movie").
		SetQueryParam("indexerIds", strconv.Itoa(indexer.ID)).
		SetResult(&result).
		Get("/api/v1/search")

	if err != nil {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, resp.Error())
		return nil, fmt.Errorf("error response from prowlarr: %v", resp.Error())
	}

	for _, torrent := range result {
		torrent.Link = strings.Replace(torrent.Link, "http://localhost:9696", j.apiURL, 1)
		torrent.InfoHash = strings.ToLower(torrent.InfoHash)
		torrent.GID = generateGID(torrent.Guid)
	}

	return result, nil
}

func (j *Prowlarr) SearchSeasonTorrents(indexer *Indexer, name string, season int) ([]*Torrent, error) {
	result := []*Torrent{}
	resp, err := j.client.
		R().
		SetQueryParam("query", fmt.Sprintf("%s{Season:%02d}", name, season)).
		SetQueryParam("categories", tvCategory).
		SetQueryParam("type", "tvsearch").
		SetQueryParam("indexerIds", strconv.Itoa(indexer.ID)).
		SetResult(&result).
		Get("/api/v1/search")

	if err != nil {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, resp.Error())
		return nil, fmt.Errorf("error response from prowlarr: %v", resp.Error())
	}

	for _, torrent := range result {
		torrent.Link = strings.Replace(torrent.Link, "http://localhost:9696", j.apiURL, 1)
		torrent.InfoHash = strings.ToLower(torrent.InfoHash)
		torrent.GID = generateGID(torrent.Guid)
	}

	return result, nil
}

func (j *Prowlarr) SearchSeriesTorrents(indexer *Indexer, name string) ([]*Torrent, error) {
	result := []*Torrent{}
	resp, err := j.client.
		R().
		SetQueryParam("query", name).
		SetQueryParam("categories", tvCategory).
		SetQueryParam("type", "tvsearch").
		SetQueryParam("indexerIds", strconv.Itoa(indexer.ID)).
		SetResult(&result).
		Get("/api/v1/search")

	if err != nil {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to search for %v from %v: %v", name, indexer.Name, resp.Error())
		return nil, fmt.Errorf("error response from prowlarr: %v", resp.Error())
	}

	for _, torrent := range result {
		normaliseTorrent(torrent, j.apiURL)
	}

	return result, nil
}

func (j *Prowlarr) FetchInfoHash(torrent *Torrent) (*Torrent, error) {
	if torrent.InfoHash != "" {
		return torrent, nil
	}

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
			log.Errorf("Unexpected magnet uri for %s, %s", torrent.Guid, torrent.Title)
			return torrent, errors.New("magnet uri is expected but not found")
		}
	}

	magnet, err := ParseMagnetUri(torrent.MagnetUri)
	if err != nil {
		return torrent, err
	}
	torrent.InfoHash = strings.ToLower(magnet.InfoHashStr())

	return torrent, nil
}

func generateGID(content string) []byte {
	h := sha1.New()
	io.WriteString(h, content)
	return h.Sum(nil)
}

func normaliseTorrent(tor *Torrent, prowlarURL string) {
	tor.Link = strings.Replace(tor.Link, "http://localhost:9696", prowlarURL, 1)
	tor.InfoHash = strings.ToLower(tor.InfoHash)
	tor.GID = generateGID(tor.Guid)
	if !strings.HasPrefix(tor.MagnetUri, "magnet") {
		if tor.Link == "" {
			tor.Link = tor.MagnetUri
		}

		if strings.HasPrefix(tor.Guid, "magnet") {
			// ThePirateBay has magnet link in Guid
			tor.MagnetUri = tor.Guid
		} else if tor.MagnetUri != "" {
			log.Errorf("Invalid magnet URI %v", tor.MagnetUri)
			tor.MagnetUri = ""
		}
	}
}
