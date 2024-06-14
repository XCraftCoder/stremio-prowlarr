package realdebrid

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2/log"
)

var (
	ErrNoTorrentFound  = errors.New("no torrent found")
	ErrTorrentNotReady = errors.New("realdebrid: torrent is not ready yet")
)

type RealDebrid struct {
	client *resty.Client
}

type File struct {
	ID       string
	FileName string `json:"filename"`
	FielSize int    `json:"filesize"`
}

type AddMagnetResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

type safeCatchedTorrentResponse map[string][]map[string]File

func (c *safeCatchedTorrentResponse) UnmarshalJSON(data []byte) error {
	mapStruct := map[string][]map[string]File(*c)
	_ = json.Unmarshal(data, &mapStruct)
	*c = mapStruct
	return nil
}

func New(apiToken string) *RealDebrid {
	client := resty.New().
		SetBaseURL("https://api.real-debrid.com/rest/1.0").
		SetHeader("Accept", "application/json").
		SetAuthScheme("Bearer").
		SetAuthToken(apiToken).
		SetDebug(true)

	return &RealDebrid{
		client: client,
	}
}

func (rd *RealDebrid) GetFiles(infoHashs []string) (map[string][]File, error) {
	result := map[string]safeCatchedTorrentResponse{}
	_, err := rd.client.R().
		SetResult(&result).
		Get("/torrents/instantAvailability/" + strings.Join(infoHashs, "/"))
	if err != nil {
		log.Errorf("Failed to get result from Debrid, err: %v", err)
		return nil, err
	}

	files := map[string][]File{}
	for infoHash, hosterFiles := range result {
		for _, variants := range hosterFiles {
			for _, variant := range variants {
				if len(variant) == 0 {
					continue
				}

				for id, f := range variant {
					newFile := f
					newFile.ID = id
					files[infoHash] = append(files[infoHash], newFile)
				}
			}
		}

	}

	return files, nil
}

func (rd *RealDebrid) GetDownloadByInfoHash(infoHash string) (string, error) {
	torrents, err := rd.getTorrents()
	if err != nil {
		return "", err
	}

	for _, torrent := range torrents {
		if torrent.Hash == infoHash {
			return rd.getDownload(&torrent)
		}
	}

	return "", ErrNoTorrentFound
}

func (rd *RealDebrid) GetDownloadByMagnetURI(magnetURI string) (string, error) {
	torrentID, err := rd.addMagnet(magnetURI)
	if err != nil {
		return "", err
	}

	torrent, err := rd.getTorrent(torrentID)
	if err != nil {
		return "", err
	}

	return rd.getDownload(torrent)
}

func (rd *RealDebrid) addMagnet(magnetUri string) (string, error) {
	result := &AddMagnetResponse{}
	_, err := rd.client.R().
		SetFormData(map[string]string{
			"magnet": magnetUri,
		}).
		SetResult(result).
		Post("/torrents/addMagnet")

	if err != nil {
		log.Errorf("Failed to select files on Debrid, err: %v", err)
		return "", err
	}

	return result.ID, nil
}

func (rd *RealDebrid) getTorrent(torrentID string) (*Torrent, error) {
	result := &Torrent{}
	_, err := rd.client.R().
		SetResult(result).
		Get("/torrents/info/" + torrentID)

	if err != nil {
		log.Errorf("Failed to fetch all torrents: %v", err)
		return nil, err
	}

	return result, nil
}

func (rd *RealDebrid) getTorrents() ([]Torrent, error) {
	result := []Torrent{}
	_, err := rd.client.R().
		SetResult(&result).
		Get("/torrents")

	if err != nil {
		log.Errorf("Failed to fetch all torrents: %v", err)
		return nil, err
	}

	return result, nil
}

func (rd *RealDebrid) getDownload(torrent *Torrent) (string, error) {
	if torrent.Status == "waiting_files_selection" {
		err := rd.selectAllFiles(torrent.ID)
		if err != nil {
			return "", err
		}

		torrent, err = rd.getTorrent(torrent.ID)
		if err != nil {
			return "", err
		}
	}

	if torrent.Status != "downloaded" {
		log.Infof("Torrent status is still %s", torrent.Status)
		return "", ErrTorrentNotReady
	}

	if len(torrent.Links) == 0 {
		return "", errors.New("not supported")
	}

	download, err := rd.generateDownload(torrent.Links[0])
	if err != nil {
		return "", err
	}

	return download, nil
}

func (rd *RealDebrid) generateDownload(hosterLink string) (string, error) {
	result := &UnrestrictedLinkResp{}
	_, err := rd.client.R().
		SetResult(&result).
		SetFormData(map[string]string{
			"link": hosterLink,
		}).
		Post("/unrestrict/link")

	if err != nil {
		log.Errorf("Failed to generate unrestricted link: %v", err)
		return "", err
	}

	return result.Download, nil
}

func (rd *RealDebrid) selectAllFiles(torrentID string) error {
	_, err := rd.client.R().
		SetFormData(map[string]string{
			"files": "all",
		}).
		Post("/torrents/selectFiles/" + torrentID)
	if err != nil {
		log.Errorf("Failed to select files on Debrid, err: %v", err)
		return err
	}

	return nil
}

type Torrent struct {
	ID          string        `json:"id"`
	Hash        string        `json:"hash"`
	Status      string        `json:"status"`
	Progress    int           `json:"progress"`
	FileName    string        `json:"filename"`
	OrgFileName string        `json:"original_filename"`
	Files       []TorrentFile `json:"files"`
	Links       []string      `json:"links"`
}

type TorrentFile struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Selected int    `json:"selected"`
	Bytes    int    `json:"bytes"`
}

type UnrestrictedLinkResp struct {
	Download string `json:"download"`
}
