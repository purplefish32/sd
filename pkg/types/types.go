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
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Pages       []Page            `json:"pages"`
	TouchScreen TouchScreenLayout `json:"touchScreen"`
}

type Instance struct {
	ID     string
	Status string
}

type Device struct {
	ID       string
	Instance string
	Type     string
	Status   string
}
