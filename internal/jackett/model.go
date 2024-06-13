package jackett

type IndexersResponse struct {
	Indexers []Indexer `xml:"indexer"`
}

type Indexer struct {
	ID    string `xml:"id,attr"`
	Title string `xml:"title"`
}
