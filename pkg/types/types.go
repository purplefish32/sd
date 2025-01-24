package types

import "encoding/json"

type Plugin interface {
	Name() string
	Init()
	GetActionTypes() []ActionType
	ValidateConfig(actionType ActionType, config json.RawMessage) error
	ExecuteAction(actionType ActionType, config json.RawMessage) error
}

type ActionType string

type Action struct {
	PluginName string          `json:"plugin_name"`
	Type       ActionType      `json:"type"`
	Config     json.RawMessage `json:"config"`
}

type ActionInstance struct {
	UUID     string  `json:"uuid"`
	Settings any     `json:"settings"`
	State    string  `json:"state"`
	States   []State `json:"states"`
	Title    string  `json:"title"`
}

type Page struct {
	ID string `json:"id"`
}

func (p Page) IsEmpty() bool {
	return p.ID == ""
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
	CurrentPage string            `json:"currentPage"`
	TouchScreen TouchScreenLayout `json:"touchScreen"`
}

func (p Profile) IsEmpty() bool {
	return p.ID == ""
}

type Instance struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (i Instance) IsEmpty() bool {
	return i.ID == ""
}

type Device struct {
	ID             string `json:"id"`
	Instance       string `json:"instance"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	CurrentProfile string `json:"currentProfile"`
}

func (d Device) IsEmpty() bool {
	return d.ID == ""
}

type State struct {
	ID        string `json:"id"`
	ImagePath string `json:"imagePath"`
}

func (s State) IsEmpty() bool {
	return s.ID == ""
}

type StateId struct {
	ID int `json:"id"`
}

func (s StateId) IsEmpty() bool {
	return s.ID == 0
}

type Button struct {
	ID       string   `json:"id"`
	UUID     string   `json:"uuid"`
	Settings Settings `json:"settings"`
	States   []State  `json:"states"`
	State    string   `json:"state"`
	Title    string   `json:"title"`
}

func (b Button) IsEmpty() bool {
	return b.ID == ""
}

type Settings struct {
	URL     string `json:"url,omitempty"`
	Text    string `json:"text,omitempty"`
	Command string `json:"command,omitempty"`
}

func (s Settings) IsEmpty() bool {
	return s.URL == "" && s.Text == "" && s.Command == ""
}
