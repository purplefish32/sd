package device

type Device struct {
	Type           string `json:"type"`
	Serial         string `json:"serial"`
	CurrentProfile string `json:"current_profile"`
}

func GetProfiles() {
	// TODO
}

func GetCurrentProfile() {
	// TODO
}

func SetCurrentProfile() {
	// TODO
}

func CreateProfile() {
	// TODO
}

func DeleteProfile() {
	// TODO
}
