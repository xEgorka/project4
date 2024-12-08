// Package models contains description of service models.
package models

import "time"

// Song describes song data.
type Song struct {
	ID          string    `json:"id" example:"ca1da5fa-50ee-4d00-82e9-d6a578419ad7"`
	Group       string    `json:"group" example:"Muse"`
	Song        string    `json:"song" example:"Supermassive Black Hole"`
	ReleaseDate time.Time `json:"release_date" format:"RFC3339" example:"2006-07-16T00:00:00Z"`
	Text        string    `json:"text" example:"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"`
	Link        string    `json:"link" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

// ResponseDetailSong describes music info api response.
type ResponseDetailSong struct {
	ReleaseDate string `json:"ReleaseDate" example:"16.07.2006"`
	Text        string `json:"text" example:"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"`
	Link        string `json:"link" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

// RequestAddSong describes song add request.
type RequestAddSong struct {
	Group string `json:"group" example:"Muse"`
	Song  string `json:"song" example:"Supermassive Black Hole"`
}

// RequestUpdateSong describes song update request.
type RequestUpdateSong struct {
	ReleaseDate time.Time `json:"release_date" format:"RFC3339" example:"2006-07-16T00:00:00Z"`
	Text        string    `json:"text" example:"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"`
	Link        string    `json:"link" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

// ResponseGetSongText describes song get text request.
type ResponseGetSongText struct {
	ID     string   `json:"id" example:"ca1da5fa-50ee-4d00-82e9-d6a578419ad7"`
	Group  string   `json:"group" example:"Muse"`
	Song   string   `json:"song" example:"Supermassive Black Hole"`
	Verses []string `json:"verses,omitempty" example:"Ooh baby don't you know I suffer?\nOoh baby can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?,Ooh\nYou set my soul alight\nOoh\nYou set my soul alight"`
	Total  int      `json:"total" example:"2"`
	Page   int      `json:"page" example:"1"`
	Size   int      `json:"size" example:"3"`
}

// ResponseGetSongs describes songs get response.
type ResponseGetSongs struct {
	Songs []Song `json:"songs"`
	Page  int    `json:"page" example:"1"`
	Size  int    `json:"size" example:"10"`
}
