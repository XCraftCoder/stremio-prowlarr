package addon

type StreamItem struct {
	URL           string               `json:"url,omitempty"`
	YoutubeID     string               `json:"ytId,omitempty"`
	InfoHash      string               `json:"infoHash,omitempty"`
	ExternalURL   string               `json:"externalUrl,omitempty"`
	Name          string               `json:"name,omitempty"`
	Description   string               `json:"description,omitempty"`
	Title         string               `json:"title,omitempty"`
	FileIndex     uint8                `json:"fileIdx,omitempty"`
	BehaviorHints *StreamBehaviorHints `json:"behaviorHints,omitempty"`
}

type StreamBehaviorHints struct {
	FileName    string `json:"filename,omitempty"`
	BingleGroup string `json:"bingeGroup,omitempty"`
	VideoSize   uint64 `json:"videoSize,omitempty"`
}
