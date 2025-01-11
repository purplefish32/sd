package plus

import (
	"encoding/json"
	"fmt"
	"sd/pkg/env"
	"sd/pkg/natsconn"
	"sd/pkg/types"
	"sd/pkg/util"

	"github.com/rs/zerolog/log"
)

// TouchScreenManager handles touch screen display state
type TouchScreenManager struct {
	plus          *Plus
	currentMode   string
	fullImage     string
	segmentImages [4]string
}

// NewTouchScreenManager creates a new touch screen manager
func NewTouchScreenManager(plus *Plus) *TouchScreenManager {
	return &TouchScreenManager{
		plus:        plus,
		currentMode: "full",
	}
}

// SetFullScreenImage sets a single image for the entire touch screen
func (tsm *TouchScreenManager) SetFullScreenImage(imagePath string) error {
	buffer, err := util.ConvertTouchScreenImageToBuffer(imagePath, ScreenWidth)
	if err != nil {
		return fmt.Errorf("failed to convert touch screen image: %w", err)
	}

	if err := tsm.plus.SetScreenImage(buffer); err != nil {
		return fmt.Errorf("failed to set touch screen image: %w", err)
	}

	tsm.currentMode = "full"
	tsm.fullImage = imagePath
	return nil
}

// SetSegmentImages sets different images for each segment
func (tsm *TouchScreenManager) SetSegmentImages(imagePaths [4]string) error {
	for segment, path := range imagePaths {
		buffer, err := util.ConvertTouchScreenImageToBuffer(path, SegmentWidth)
		if err != nil {
			return fmt.Errorf("failed to convert segment %d image: %w", segment+1, err)
		}
		if err := tsm.plus.SetScreenSegment(segment+1, buffer); err != nil {
			return fmt.Errorf("failed to set segment %d: %w", segment+1, err)
		}
	}

	tsm.currentMode = "segments"
	tsm.segmentImages = imagePaths
	return nil
}

// BlankTouchScreen sets the entire touch screen to black
func (tsm *TouchScreenManager) BlankTouchScreen() error {
	assetPath := env.Get("ASSET_PATH", "")
	return tsm.SetFullScreenImage(assetPath + "images/black.png")
}

// UpdateFromProfile updates the touch screen display based on profile settings
func (tsm *TouchScreenManager) UpdateFromProfile(profile *types.Profile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}

	switch profile.TouchScreen.Mode {
	case "full":
		return tsm.SetFullScreenImage(profile.TouchScreen.FullImage)
	case "segments":
		return tsm.SetSegmentImages(profile.TouchScreen.Segments)
	default:
		return tsm.BlankTouchScreen()
	}
}

// WatchProfileChanges watches for profile changes and updates the display
func (tsm *TouchScreenManager) WatchProfileChanges(instanceID string) {
	_, kv := natsconn.GetNATSConn()

	pattern := fmt.Sprintf("instances.%s.devices.%s.profiles.*",
		instanceID, tsm.plus.device.Serial)

	watcher, err := kv.Watch(pattern)
	if err != nil {
		log.Error().Err(err).Msg("Failed to watch profiles")
		return
	}

	for update := range watcher.Updates() {
		if update == nil {
			continue
		}

		var profile types.Profile
		if err := json.Unmarshal(update.Value(), &profile); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal profile")
			continue
		}

		if err := tsm.UpdateFromProfile(&profile); err != nil {
			log.Error().Err(err).Msg("Failed to update touch screen")
		}
	}
}
