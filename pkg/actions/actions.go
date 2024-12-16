package actions

import "github.com/nats-io/nats.go"

type Action struct {
	Type   string `json:"type"` // Key or Dial
	Params string `json:"params"`
}

type State struct {
	Id        string `json:"id"`
	ImagePath string `json:"imagePath"`
}
type ActionInstance struct {
	UUID     string  `json:"uuid"`
	Settings any     `json:"settings"`
	State    string  `json:"state"`
	States   []State `json:"states"`
	Title    string  `json:"title"`
}

func Update(kv nats.KeyValue, actionType string, params struct{}) {
	kv.PutString("", "") // Try storing as byte[]
}

func Delete(kv nats.KeyValue, actionType string, params struct{}) {
	kv.Delete("") // Try storing as byte[]
}
