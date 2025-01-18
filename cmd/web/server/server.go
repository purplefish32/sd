package server

import (
	"bytes"
	"context"
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

	"sd/cmd/web/views/partials"
	"sd/pkg/natsconn"
	"sd/pkg/types"
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
		instances, err := s.getInstances()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.HomePage(instances).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}", func(w http.ResponseWriter, r *http.Request) {
		instances, err := s.getInstances()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := s.getDevices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.InstancePage(instances, devices).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceId}/device/{deviceId}", func(w http.ResponseWriter, r *http.Request) {
		instances, err := s.getInstances()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := s.getDevices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.DevicePage(instances, devices).Render(r.Context(), w)
	})

	// HTMX Routes
	s.router.Get("/partials/instance-card-list", s.handleInstanceCardList)
	s.router.Get("/partials/device-card-list", s.handleDeviceCardList)

	s.router.Get("/partials/button", s.handleButton)

	// s.router.Get("/instances/{id}/devices", func(w http.ResponseWriter, r *http.Request) {
	// 	instanceID := chi.URLParam(r, "id")
	// 	s.log.Info().Str("instance", instanceID).Msg("Loading instance devices")

	// 	devices, err := s.getDevicesForInstance(instanceID)
	// 	if err != nil {
	// 		s.log.Error().E<!-- Left Panel - Instance List -->
	// 	}

	// 	// Render the device list with SSE support
	// 	components.DeviceList(devices).Render(r.Context(), w)
	// })

	// Add SSE endpoint for device updates
	s.router.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info().
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("New SSE connection")

		// 1. Set proper headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*") // Add CORS if needed

		// 2. Create done channel with buffer to prevent goroutine leak
		done := make(chan bool, 1)
		defer close(done)

		// 3. Create watcher with proper error handling
		watcher, err := s.kv.Watch("instances.*.devices.*")
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to create KV watcher")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer watcher.Stop()

		// 4. Send initial device list
		devices, err := s.getDevices()
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to get initial devices")
		} else {
			if err := s.sendDeviceList(w, r.Context(), devices); err != nil {
				s.log.Error().Err(err).Msg("Failed to send initial device list")
				return
			}
		}

		// 5. Handle client disconnection
		go func() {
			<-r.Context().Done()
			done <- true
		}()

		// 6. Main event loop
		for {
			select {
			case <-done:
				return
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				select {
				case <-done:
					return
				default:
					devices, err := s.getDevices()
					if err != nil {
						s.log.Error().Err(err).Msg("Failed to get devices")
						continue
					}

					if err := s.sendDeviceList(w, r.Context(), devices); err != nil {
						if err != context.Canceled {
							s.log.Error().Err(err).Msg("Failed to send device list")
						}
						return
					}
				}
			}
		}
	})
}

// Move sendDeviceList outside setupRoutes
func (s *Server) sendDeviceList(w http.ResponseWriter, ctx context.Context, devices []types.Device) error {
	var buf bytes.Buffer
	if err := partials.DeviceCardList(devices).Render(ctx, &buf); err != nil {
		return err
	}

	fmt.Fprintf(w, "event: DeviceCardListUpdate\n")
	fmt.Fprintf(w, "data: %s\n\n", buf.String())

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func (s *Server) getDevices() ([]types.Device, error) {
	keyList, err := s.kv.ListKeys()
	if err != nil {
		return nil, err
	}

	devices := make([]types.Device, 0)
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

			devices = append(devices, types.Device{
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

func (s *Server) handleButton(w http.ResponseWriter, r *http.Request) {
	s.log.Info().Msg("TODO handle button") // TODO: Implement this
}

func (s *Server) handleDeviceCardList(w http.ResponseWriter, r *http.Request) {
	s.log.Info().Msg("Handling device list request")

	devices, err := s.getDevices()

	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get devices")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info().Interface("devices", devices).Msg("Found devices")
	partials.DeviceCardList(devices).Render(r.Context(), w)
}

// TODO move this to the instance package.
func (s *Server) getInstances() ([]types.Instance, error) {
	keys, err := s.kv.Keys()
	if err != nil {
		return nil, err
	}

	instances := make([]types.Instance, 0)
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

			instances = append(instances, types.Instance{
				ID:     instanceID,
				Status: "Connected", // TODO: Get actual status
			})
			seen[instanceID] = true
		}
	}

	return instances, nil
}

func (s *Server) handleInstanceCardList(w http.ResponseWriter, r *http.Request) {
	s.log.Info().Msg("Handling instance list request")

	instances, err := s.getInstances()
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to get instances")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Info().Interface("instances", instances).Msg("Found instances")
	partials.InstanceCardList(instances).Render(r.Context(), w)
}

func (s *Server) getDevicesForInstance(instanceID string) ([]types.Device, error) {
	keyList, err := s.kv.ListKeys()
	if err != nil {
		return nil, err
	}

	devices := make([]types.Device, 0)
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

		devices = append(devices, types.Device{
			ID:       deviceID,
			Instance: instanceID,
			Type:     deviceInfo.Type,
			Status:   deviceInfo.Status,
		})
		seen[deviceID] = true
	}

	return devices, nil
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
