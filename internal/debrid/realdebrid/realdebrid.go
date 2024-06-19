package realdebrid

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2/log"
)

var (
	ErrNoTorrentFound  = errors.New("no torrent found")
	ErrNoFileFound     = errors.New("realdebrid: not file found")
	ErrTorrentNotReady = errors.New("realdebrid: torrent is not ready yet")
)

type RealDebrid struct {
	client *resty.Client
}

type File struct {
	ID       string
	FileName string `json:"filename"`
	FileSize uint64 `json:"filesize"`
}

type AddMagnetResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

type safeCatchedTorrentResponse map[string][]map[string]*File

func (c *safeCatchedTorrentResponse) UnmarshalJSON(data []byte) error {
	mapStruct := map[string][]map[string]*File(*c)
	_ = json.Unmarshal(data, &mapStruct)
	*c = mapStruct
	return nil
}

func New(apiToken string) *RealDebrid {
	client := resty.New().
		SetBaseURL("https://api.real-debrid.com/rest/1.0").
		SetHeader("Accept", "application/json").
		SetAuthScheme("Bearer").
		SetError(ErrorResponse{}).
		SetAuthToken(apiToken)

	return &RealDebrid{
		client: client,
	}
}

func (rd *RealDebrid) GetFiles(infoHashs []string) (map[string][]*File, error) {
	result := map[string]safeCatchedTorrentResponse{}
	resp, err := rd.client.R().
		SetResult(&result).
		Get("/torrents/instantAvailability/" + strings.Join(infoHashs, "/"))
	if err != nil {
		log.Errorf("Failed to get result from Debrid, err: %v", err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to get result from Debrid, err: %v", resp.Error())
		return nil, resp.Error().(error)
	}

	files := map[string][]*File{}
	found := map[string]bool{}
	for infoHash, hosterFiles := range result {
		for _, variants := range hosterFiles {
			for _, variant := range variants {
				for id, f := range variant {
					if !found[id] {
						newFile := f
						newFile.ID = id
						files[infoHash] = append(files[infoHash], newFile)
						found[id] = true
					}
				}
			}
		}

	}

	return files, nil
}

func (rd *RealDebrid) GetDownloadByMagnetURI(infoHash string, magnetURI string, fileID string, ipAddress string) (string, error) {
	download, err := rd.getDownloadByInfoHash(infoHash, fileID, ipAddress)
	if err == nil {
		return download, nil
	}

	if err != ErrNoTorrentFound {
		return "", err
	}

	torrentID, err := rd.addMagnet(magnetURI, ipAddress)
	if err != nil {
		return "", err
	}

	torrent, err := rd.getTorrent(torrentID)
	if err != nil {
		return "", err
	}

	return rd.getDownload(torrent, fileID, ipAddress)
}

func (rd *RealDebrid) getDownloadByInfoHash(infoHash, fileID, ipAddress string) (string, error) {
	torrents, err := rd.getTorrents()
	if err != nil {
		return "", err
	}

	for _, torrent := range torrents {
		if torrent.Hash == infoHash {
			download, err := rd.getDownload(&torrent, fileID, ipAddress)
			if err == nil {
				return download, err
			}

			if err != ErrNoFileFound {
				return "", err
			}
		}
	}

	return "", ErrNoTorrentFound
}

func (rd *RealDebrid) addMagnet(magnetUri string, ipAddress string) (string, error) {
	result := &AddMagnetResponse{}
	formData := withIPAddress(map[string]string{
		"magnet": magnetUri,
	}, ipAddress)

	resp, err := rd.client.R().
		SetFormData(formData).
		SetResult(result).
		Post("/torrents/addMagnet")

	if err != nil {
		log.Errorf("Failed to select files on Debrid, err: %v", err)
		return "", err
	}

	if resp.IsError() {
		log.Errorf("Failed to get result from Debrid, err: %v", resp.Error())
		return "", resp.Error().(error)
	}

	return result.ID, nil
}

func (rd *RealDebrid) getTorrent(torrentID string) (*Torrent, error) {
	result := &Torrent{}
	resp, err := rd.client.R().
		SetResult(result).
		Get("/torrents/info/" + torrentID)

	if err != nil {
		log.Errorf("Failed to fetch all torrents: %v", err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to get result from Debrid, err: %v", resp.Error())
		return nil, resp.Error().(error)
	}

	return result, nil
}

func (rd *RealDebrid) getTorrents() ([]Torrent, error) {
	result := []Torrent{}
	resp, err := rd.client.R().
		SetResult(&result).
		SetQueryParam("limit", "200").
		SetQueryParam("filter", "active").
		Get("/torrents")

	if err != nil {
		log.Errorf("Failed to fetch all torrents: %v", err)
		return nil, err
	}

	if resp.IsError() {
		log.Errorf("Failed to get torrents from Debrid, err: %v", resp.Error())
		return nil, resp.Error().(error)
	}

	return result, nil
}

func (rd *RealDebrid) getDownload(torrent *Torrent, fileID string, ipAddress string) (string, error) {
	linkIndex := getIndexOfLinkForFile(torrent, fileID)
	if torrent.Status == "waiting_files_selection" || linkIndex == -1 {
		err := rd.selectFileToDownload(torrent.ID, ipAddress)
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

	linkIndex = getIndexOfLinkForFile(torrent, fileID)
	if linkIndex == -1 {
		return "", ErrNoFileFound
	}

	if len(torrent.Links) == 0 || len(torrent.Links) <= linkIndex {
		log.Infof("Invalid torrent link: %d, len: %d", linkIndex, len(torrent.Links))
		return "", errors.New("not supported")
	}

	download, err := rd.generateDownload(torrent.Links[linkIndex], ipAddress)
	if err != nil {
		return "", err
	}

	return download, nil
}

func (rd *RealDebrid) generateDownload(hosterLink string, ipAddress string) (string, error) {
	result := &UnrestrictedLinkResp{}
	formData := withIPAddress(map[string]string{
		"link": hosterLink,
	}, ipAddress)

	resp, err := rd.client.R().
		SetResult(&result).
		SetFormData(formData).
		Post("/unrestrict/link")

	if err != nil {
		log.Errorf("Failed to generate unrestricted link: %v", err)
		return "", err
	}

	if resp.IsError() {
		log.Errorf("Failed to generate download link from Debrid, err: %v", resp.Error())
		return "", resp.Error().(error)
	}

	return result.Download, nil
}

func (rd *RealDebrid) selectFileToDownload(torrentID, ipAddress string) error {
	formData := withIPAddress(map[string]string{
		"files": "all",
	}, ipAddress)

	resp, err := rd.client.R().
		SetFormData(formData).
		Post("/torrents/selectFiles/" + torrentID)
	if err != nil {
		log.Errorf("Failed to select files on Debrid, err: %v", err)
		return err
	}

	if resp.IsError() {
		log.Errorf("Failed to select files from Debrid, err: %v", resp.Error())
		return resp.Error().(error)
	}

	return nil
}

func getIndexOfLinkForFile(torrent *Torrent, fileID string) int {
	index := 0
	for _, f := range torrent.Files {
		if fmt.Sprint(f.ID) == fileID {
			if f.Selected > 0 {
				return index
			}

			return -1
		}

		if f.Selected > 0 {
			index++
		}
	}

	return -1
}

type Torrent struct {
	ID          string        `json:"id"`
	Hash        string        `json:"hash"`
	Status      string        `json:"status"`
	Progress    float64       `json:"progress"`
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

type ErrorResponse struct {
	ErrTxt    string `json:"error"`
	ErrorCode int    `json:"error_code"`
}

func (er ErrorResponse) Error() string {
	return fmt.Sprintf("[%s,%d]", er.ErrTxt, er.ErrorCode)
}

func withIPAddress(formData map[string]string, ipAddress string) map[string]string {
	if ipAddress != "" {
		formData["ip"] = ipAddress
	}

	return formData
}
