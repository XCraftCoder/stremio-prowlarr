package jackett

import (
	"encoding/hex"
)

type TorrentID []byte

type IndexersResponse struct {
	Indexers []Indexer `xml:"indexer"`
}

type Indexer struct {
	ID    string `xml:"id,attr"`
	Title string `xml:"title"`
}

type Torrent struct {
	GID       TorrentID
	Title     string   `json:"Title"`
	Guid      string   `json:"Guid"`
	Seeders   uint     `json:"Seeders"`
	Size      uint     `json:"Size"`
	Imdb      uint     `json:"Imdb"`
	TMDb      uint     `json:"TMDb"`
	TVDBId    uint     `json:"TVDBId"`
	Link      string   `json:"Link"`
	MagnetUri string   `json:"MagnetUri"`
	InfoHash  string   `json:"InfoHash"`
	Year      uint     `json:"Year"`
	Languages []string `json:"Languages"`
	Subs      []string `json:"Subs"`
	Peers     uint     `json:"Peers"`
	Files     uint     `json:"Files"`
}

type TorrentsResponse struct {
	Torrents []Torrent `json:"Results"`
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
