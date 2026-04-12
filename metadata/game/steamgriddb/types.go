package steamgriddb

// Game represents a game from SteamGridDB.
type Game struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Types    []string `json:"types"`
	Verified bool     `json:"verified"`
}

// Image represents an image (grid, hero, logo, or icon) from SteamGridDB.
type Image struct {
	ID        int    `json:"id"`
	Score     int    `json:"score"`
	Style     string `json:"style"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	NSFW      bool   `json:"nsfw"`
	Humor     bool   `json:"humor"`
	Notes     string `json:"notes"`
	Mime      string `json:"mime"`
	Language  string `json:"language"`
	URL       string `json:"url"`
	Thumb     string `json:"thumb"`
	Lock      bool   `json:"lock"`
	Epilepsy  bool   `json:"epilepsy"`
	Upvotes   int    `json:"upvotes"`
	Downvotes int    `json:"downvotes"`
	Author    Author `json:"author"`
}

// Author represents the user who uploaded an image.
type Author struct {
	Name    string `json:"name"`
	Steam64 string `json:"steam64"`
	Avatar  string `json:"avatar"`
}

// SearchResult represents a game search result.
type SearchResult struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Types    []string `json:"types"`
	Verified bool     `json:"verified"`
}

// response wraps the standard SteamGridDB API response.
type response[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}
