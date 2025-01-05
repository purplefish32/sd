package actions

type HotkeyConfig struct {
	Keys []string `json:"keys"` // e.g., ["ctrl", "shift", "a"]
}

func (h HotkeyConfig) Execute() error {
	// Implementation for pressing hotkeys
	return nil
}
