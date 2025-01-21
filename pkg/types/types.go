package types

type Page struct {
	ID string `json:"id"`
}

type TouchScreenLayout struct {
	Mode      string    `json:"mode"`
	FullImage string    `json:"fullImage"`
	Segments  [4]string `json:"segments"`
}

type Profile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Pages []Page `json:"pages"`
}

type Instance struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Device struct {
	ID       string `json:"id"`
	Instance string `json:"instance"`
	Type     string `json:"type"`
	Status   string `json:"status"`
}
