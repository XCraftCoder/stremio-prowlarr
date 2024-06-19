package addon

// ContentType refers to https://github.com/Stremio/stremio-addon-sdk/blob/master/docs/api/responses/content.types.md
type ContentType string

const (
	ContentTypeMovie  ContentType = "movie"
	ContentTypeSeries ContentType = "series"
)

// Resource refers to https://github.com/Stremio/stremio-addon-sdk/blob/master/docs/api/responses/manifest.md#filtering-properties
type Resource string

const (
	ResourceStream Resource = "stream"
)

type Manifest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`

	ResourceItems []ResourceItem `json:"resources,omitempty"`

	Types    []ContentType `json:"types"`
	Catalogs []CatalogItem `json:"catalogs,omitempty"`

	IDPrefixes    []string       `json:"idPrefixes,omitempty"`
	Background    string         `json:"background,omitempty"`
	Logo          string         `json:"logo,omitempty"`
	ContactEmail  string         `json:"contactEmail,omitempty"`
	BehaviorHints *BehaviorHints `json:"behaviorHints,omitempty"`
}

type ResourceItem struct {
	Name  Resource      `json:"name"`
	Types []ContentType `json:"types"`

	IDPrefixes []string `json:"idPrefixes,omitempty"`
}

type BehaviorHints struct {
	Adult                 bool `json:"adult,omitempty"`
	P2P                   bool `json:"p2p,omitempty"`
	Configurable          bool `json:"configurable,omitempty"`
	ConfigurationRequired bool `json:"configurationRequired,omitempty"`
}

// CatalogItem represents a catalog.
type CatalogItem struct {
	Type ContentType `json:"type"`
	ID   string      `json:"id"`
	Name string      `json:"name"`

	Extra []ExtraItem `json:"extra,omitempty"`
}

type ExtraItem struct {
	Name string `json:"name"`

	IsRequired   bool     `json:"isRequired,omitempty"`
	Options      []string `json:"options,omitempty"`
	OptionsLimit int      `json:"optionsLimit,omitempty"`
}
