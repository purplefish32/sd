package types

type Page struct {
	ID string `json:"id"`
}

type TouchScreenLayout struct {
	Mode      string    `json:"mode"`      // "full" or "segments"
	FullImage string    `json:"fullImage"` // Path for full screen image
	Segments  [4]string `json:"segments"`  // Paths for segment images
}

type Profile struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Pages       []Page            `json:"pages"`
	TouchScreen TouchScreenLayout `json:"touchScreen"`
}
