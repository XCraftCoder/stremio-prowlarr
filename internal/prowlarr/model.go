package prowlarr

import (
	"encoding/hex"
)

type TorrentID []byte

type Indexer struct {
	ID           int                 `json:"id"`
	Name         string              `json:"name"`
	SortName     string              `json:"sortName"`
	Enable       bool                `json:"enable"`
	Capabilities IndexerCapabilities `json:"capabilities"`
}

type IndexerCapabilities struct {
	LimitMax      int `json:"limitsMax"`
	LimitDefaults int `json:"limitsDefault"`
}

type Torrent struct {
	GID       TorrentID
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	FileName  string   `json:"fileName"`
	Guid      string   `json:"guid"`
	Seeders   uint     `json:"seeders"`
	Size      uint     `json:"size"`
	Imdb      uint     `json:"imdbId"`
	TMDb      uint     `json:"TMDb"`
	TVDBId    uint     `json:"TVDBId"`
	Link      string   `json:"downloadUrl"`
	MagnetUri string   `json:"magnetUrl"`
	InfoHash  string   `json:"infoHash"`
	Year      uint     `json:"Year"`
	Languages []string `json:"Languages"`
	Subs      []string `json:"Subs"`
	Peers     uint     `json:"Peers"`
	Files     uint     `json:"files"`
}

type RSSItem struct {
	Channel ChannelItem `xml:"channel"`
}

type ChannelItem struct {
	Items []Torrent `xml:"item"`
}

func (t TorrentID) ToString() string {
	return hex.EncodeToString(t)
}

func TorrentIDFromString(encoded string) (TorrentID, error) {
	return hex.DecodeString(encoded)
}
