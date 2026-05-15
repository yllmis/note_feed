package search

type Article struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
