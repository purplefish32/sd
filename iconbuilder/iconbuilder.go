package iconbuilder

type IconRequest struct {
	BackgroundColor string `json:"background_color"` // Hex, e.g., "#FF0000"
	IconPath        string `json:"icon_path"`        // Path to icon file
	Text            string `json:"text"`             // Text to render
}

type IconResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []byte `json:"data,omitempty"` // PNG binary (optional)
}
