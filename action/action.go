package action

import "github.com/nats-io/nats.go"

type Action struct {
	Type   string `json:"type"` // Key or Dial
	Params string `json:"params"`
}

// func Create(kv nats.KeyValue, actionType string, params struct{}) {
// 	kv.Create("", _) // Try storing as byte[]
// }

func Update(kv nats.KeyValue, actionType string, params struct{}) {
	kv.PutString("", "") // Try storing as byte[]
}

func Delete(kv nats.KeyValue, actionType string, params struct{}) {
	kv.Delete("") // Try storing as byte[]
}
