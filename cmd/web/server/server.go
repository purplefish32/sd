package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/cmd/web/views/components"
	"sd/cmd/web/views/pages"
	"sd/pkg/natsconn"
)

type Server struct {
	router *chi.Mux
	log    zerolog.Logger
	nc     *nats.Conn
	kv     nats.KeyValue
}

// Add these constants for device types
const (
	DeviceTypeXL    = "xl"
	DeviceTypePlus  = "plus"
	DeviceTypePedal = "pedal"

	// USB IDs
	VendorIDElgato = 0x0fd9
	ProductIDXL    = 0x006c // Stream Deck XL
	ProductIDPlus  = 0x0084 // Stream Deck +
	ProductIDPedal = 0x0086 // Stream Deck Pedal
)

// DeviceInfo represents the device data stored in NATS KV
type DeviceInfo struct {
	Type      string    `json:"type"`       // xl, plus, pedal
	CreatedAt time.Time `json:"created_at"` // When the device was first seen
	UpdatedAt time.Time `json:"updated_at"` // Last time device was seen
	Status    string    `json:"status"`     // connected, disconnected
}

// DetermineDeviceType returns the device type based on USB product ID
func DetermineDeviceType(productID uint16) string {
	switch productID {
	case ProductIDXL:
		return DeviceTypeXL
	case ProductIDPlus:
		return DeviceTypePlus
	case ProductIDPedal:
		return DeviceTypePedal
	default:
		return "unknown"
	}
}

func NewServer() *Server {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Connect to NATS
	nc, kv := natsconn.GetNATSConn()

	log.Info().Interface("nc", nc).Msg("NATS Connection")
	log.Info().Interface("kv", kv).Msg("NATS KV")

	if nc == nil || kv == nil {
		log.Error().Msg("Failed to connect to NATS")
	} else {
		log.Info().Msg("Connected to NATS successfully")
	}

	s := &Server{
		router: r,
		log:    log.With().Str("component", "web").Logger(),
		nc:     nc,
		kv:     kv,
	}

	// Setup routes
	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Routes
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pages.Home().Render(r.Context(), w)
	})

	// HTMX Routes
	s.router.Get("/devices/list", s.handleDeviceList)
	s.router.Get("/instances/list", s.handleInstanceList)
	s.router.Get("/instances/{id}/devices", func(w http.ResponseWriter, r *http.Request) {
		instanceID := chi.URLParam(r, "id")
		s.log.Info().Str("instance", instanceID).Msg("Loading instance devices")

		devices, err := s.getDevicesForInstance(instanceID)
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to get devices")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Render the device list with SSE support
		components.DeviceList(devices).Render(r.Context(), w)
	})

	// Add SSE endpoint for device updates
	s.router.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		s.log.Info().
			Str("remote_addr", r.RemoteAddr).
			Msg("New SSE connection established")

		s.log.Info().Msg("Client connected to SSE stream")

		watcher, err := s.kv.Watch("instances.*.devices.*")
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to create KV watcher")
			return
		}
		defer watcher.Stop()

		lastEvents := make(map[string]string)

		for {
			select {
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				var deviceInfo DeviceInfo

				s.log.Info().Interface("entry_key", entry.Key()).Msg("Received entry")

				if err := json.Unmarshal(entry.Value(), &deviceInfo); err != nil {
					s.log.Error().Err(err).Msg("Failed to parse device info")
					continue
				}

				parts := strings.Split(entry.Key(), ".")

				if len(parts) < 4 {
					continue
				}

				deviceID := parts[3]
				eventKey := fmt.Sprintf("%s:%s", deviceID, deviceInfo.Status)

				if lastEvent, ok := lastEvents[deviceID]; ok && lastEvent == eventKey {
					continue
				}

				lastEvents[deviceID] = eventKey

				// Create device component
				device := components.Device{
					ID:       deviceID,
					Type:     deviceInfo.Type,
					Status:   deviceInfo.Status,
					Instance: parts[1],
				}

				// Render the device status component
				var buf bytes.Buffer
				if err := components.DeviceCard(device).Render(r.Context(), &buf); err != nil {
					s.log.Error().Err(err).Msg("Failed to render device status")
					continue
				}

				// Send event with device-specific ID
				fmt.Fprintf(w, "event: device-update-%s\n", deviceID)
				fmt.Fprintf(w, "data: %s\n", buf.String())
				fmt.Fprintf(w, "\n")

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}

			case <-r.Context().Done():
				return
			}
		}
	})

	s.router.Get("/devices/{id}/config", func(w http.ResponseWriter, r *http.Request) {
		deviceID := chi.URLParam(r, "id")

		// Find instance ID for this device
		devices, err := s.getDevices()
		if err != nil {
			http.Error(w, "Failed to get devices", http.StatusInternalServerError)
			return
		}

		var instanceID string
		for _, d := range devices {
			if d.ID == deviceID {
				instanceID = d.Instance
				break
			}
		}

		if instanceID == "" {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		deviceInfo, err := s.getDeviceInfo(instanceID, deviceID)
		if err != nil {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}

		switch deviceInfo.Type {
		case "xl":
			components.StreamDeckXL(deviceID).Render(r.Context(), w)
		case "plus":
			components.StreamDeckPlus(deviceID).Render(r.Context(), w)
		case "pedal":
			components.StreamDeckPedal(deviceID).Render(r.Context(), w)
		default:
			http.Error(w, "Unsupported device type", http.StatusBadRequest)
		}
	})
}

func (s *Server) getDevices() ([]components.Device, error) {
	keyList, err := s.kv.ListKeys()
	if err != nil {
		return nil, err
	}

	devices := make([]components.Device, 0)
	seen := make(map[string]bool)

	for key := range keyList.Keys() {
		if strings.Contains(key, "devices") {
			parts := strings.Split(key, ".")
			if len(parts) < 4 {
				continue
			}

			deviceID := parts[3]
			if seen[deviceID] {
				continue
			}

			entry, err := s.kv.Get(key)
			if err != nil {
				s.log.Warn().Err(err).Str("key", key).Msg("Skipping invalid device entry")
				continue
			}

			var deviceInfo DeviceInfo
			if err := json.Unmarshal(entry.Value(), &deviceInfo); err != nil {
				s.log.Warn().Err(err).Str("key", key).Msg("Skipping malformed device data")
				continue
			}

			devices = append(devices, components.Device{
				ID:       deviceID,
				Instance: parts[1],
				Type:     deviceInfo.Type,
				Status:   deviceInfo.Status,
			})
			seen[deviceID] = true
		}
	}

	return devices, nil
}

func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.getDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func (s *Server) handleDeviceList(w http.ResponseWriter, r *http.Request) {
	s.log.Info().Msg("Handling device list request")

	devices, err := s.getDevices()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get devices")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info().Interface("devices", devices).Msg("Found devices")
	components.DeviceList(devices).Render(r.Context(), w)
}

func (s *Server) getInstances() ([]components.Instance, error) {
	keys, err := s.kv.Keys()
	if err != nil {
		return nil, err
	}

	instances := make([]components.Instance, 0)
	seen := make(map[string]bool)

	for _, key := range keys {
		if strings.HasPrefix(key, "instances.") {
			parts := strings.Split(key, ".")
			if len(parts) < 2 {
				continue
			}

			instanceID := parts[1]
			if seen[instanceID] {
				continue // Skip duplicates
			}

			instances = append(instances, components.Instance{
				ID:     instanceID,
				Status: "Connected", // TODO: Get actual status
			})
			seen[instanceID] = true
		}
	}

	return instances, nil
}

func (s *Server) handleInstanceList(w http.ResponseWriter, r *http.Request) {
	s.log.Info().Msg("Handling instance list request")

	instances, err := s.getInstances()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get instances")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info().Interface("instances", instances).Msg("Found instances")
	components.InstanceList(instances).Render(r.Context(), w)
}

func (s *Server) handleInstanceDevices(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "id")
	s.log.Info().Str("instance", instanceID).Msg("Loading instance devices")

	devices, err := s.getDevicesForInstance(instanceID)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get devices")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	components.DeviceList(devices).Render(r.Context(), w)
}

func (s *Server) getDevicesForInstance(instanceID string) ([]components.Device, error) {
	keyList, err := s.kv.ListKeys()
	if err != nil {
		return nil, err
	}

	devices := make([]components.Device, 0)
	seen := make(map[string]bool)

	for key := range keyList.Keys() {
		// Only match direct device keys, not sub-keys like profiles
		parts := strings.Split(key, ".")
		if len(parts) != 4 || parts[0] != "instances" || parts[2] != "devices" {
			continue
		}
		if parts[1] != instanceID {
			continue
		}

		deviceID := parts[3]
		if seen[deviceID] {
			continue
		}

		entry, err := s.kv.Get(key)
		if err != nil {
			s.log.Warn().Err(err).Str("key", key).Msg("Skipping invalid device entry")
			continue
		}

		var deviceInfo DeviceInfo
		if err := json.Unmarshal(entry.Value(), &deviceInfo); err != nil {
			s.log.Warn().Err(err).Str("key", key).Msg("Skipping malformed device data")
			continue
		}

		devices = append(devices, components.Device{
			ID:       deviceID,
			Instance: instanceID,
			Type:     deviceInfo.Type,
			Status:   deviceInfo.Status,
		})
		seen[deviceID] = true
	}

	return devices, nil
}

// Store a new device with auto-detection
func (s *Server) storeDevice(instanceID string, deviceID string, productID uint16) error {
	deviceType := DetermineDeviceType(productID)
	if deviceType == "unknown" {
		return fmt.Errorf("unknown device type for product ID: %x", productID)
	}

	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)

	info := DeviceInfo{
		Type:      deviceType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "connected",
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = s.kv.Put(key, data)
	if err != nil {
		return fmt.Errorf("failed to store device info: %w", err)
	}

	s.log.Info().
		Str("instance", instanceID).
		Str("device", deviceID).
		Str("type", deviceType).
		Msg("Stored new device")

	return nil
}

// Get device info
func (s *Server) getDeviceInfo(instanceID, deviceID string) (*DeviceInfo, error) {
	key := fmt.Sprintf("instances.%s.devices.%s", instanceID, deviceID)

	entry, err := s.kv.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	var info DeviceInfo
	if err := json.Unmarshal(entry.Value(), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return &info, nil
}

func (s *Server) Start() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	s.log.Info().Msgf("Starting server on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}

func (s *Server) Close() {
}
