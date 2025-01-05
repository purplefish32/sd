package types

type Page struct {
	ID string `json:"id"`
}

type Profile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Pages []Page `json:"pages"`
}
