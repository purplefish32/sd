package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"sd/cmd/web/views/partials"
	"sd/pkg/buttons"
	"sd/pkg/devices"
	"sd/pkg/instance"
	"sd/pkg/natsconn"
	"sd/pkg/profiles"
	"sd/pkg/types"
)

type Server struct {
	router *chi.Mux
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
		instances, err := instance.GetInstances()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.HomePage(instances).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}", func(w http.ResponseWriter, r *http.Request) {
		instances, err := instance.GetInstances()
		instanceID := chi.URLParam(r, "instanceID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := devices.GetDevices(instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.InstancePage(instances, devices).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}/device/{deviceID}", func(w http.ResponseWriter, r *http.Request) {
		instances, err := instance.GetInstances()
		instanceID := chi.URLParam(r, "instanceID")
		deviceID := chi.URLParam(r, "deviceID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := devices.GetDevices(instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		profiles, err := profiles.GetProfiles(instanceID, deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pages, err := s.getPages()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.DevicePage(instances, devices, profiles, pages, instanceID, deviceID).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}/device/{deviceID}/profile/{profileID}", func(w http.ResponseWriter, r *http.Request) {
		instances, err := instance.GetInstances()
		instanceID := chi.URLParam(r, "instanceID")
		deviceID := chi.URLParam(r, "deviceID")
		profileID := chi.URLParam(r, "profileID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		devices, err := devices.GetDevices(instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		profiles, err := profiles.GetProfiles(instanceID, deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pages, err := s.getPages()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		partials.ProfilePage(instances, devices, profiles, pages, instanceID, deviceID, profileID).Render(r.Context(), w)
	})

	// HTMX Routes
	s.router.Get("/partials/instance-card-list", s.handleInstanceCardList)
	s.router.Get("/partials/{instanceId}/device-card-list", s.handleDeviceCardList)
	s.router.Get("/partials/button/{instanceId}/{deviceId}/{profileId}/{pageId}/{buttonId}", s.handleButton)
	s.router.Post("/partials/button/{instanceId}/{deviceId}/{profileId}/{pageId}/{buttonId}", s.handleButtonPress)

	// Add SSE endpoint for device updates
	s.router.Get("/stream/{instanceId}", func(w http.ResponseWriter, r *http.Request) {
		instanceID := chi.URLParam(r, "instanceId")
		log.Info().
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
		watcher, err := s.kv.Watch("instances." + instanceID + ".devices.*")

		if err != nil {
			log.Error().Err(err).Msg("Failed to create KV watcher")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer watcher.Stop()

		// 4. Send initial device list
		d, err := devices.GetDevices(instanceID)

		log.Info().Interface("devices", d).Msg("Initial devices")

		if err != nil {
			log.Error().Err(err).Msg("Failed to get initial devices")
		} else {
			if err := s.sendDeviceList(w, r.Context(), instanceID, d); err != nil {
				log.Error().Err(err).Msg("Failed to send initial device list")
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
					d, err := devices.GetDevices(instanceID)
					if err != nil {
						log.Error().Err(err).Msg("Failed to get devices")
						continue
					}

					if err := s.sendDeviceList(w, r.Context(), instanceID, d); err != nil {
						if err != context.Canceled {
							log.Error().Err(err).Msg("Failed to send device list")
						}
						return
					}
				}
			}
		}
	})
}

// Move sendDeviceList outside setupRoutes
func (s *Server) sendDeviceList(w http.ResponseWriter, ctx context.Context, instanceID string, devices []types.Device) error {
	var buf bytes.Buffer
	if err := partials.DeviceCardList(instanceID, devices).Render(ctx, &buf); err != nil {
		return err
	}

	fmt.Fprintf(w, "event: DeviceCardListUpdate\n")
	fmt.Fprintf(w, "data: %s\n\n", buf.String())

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func (s *Server) getPages() ([]types.Page, error) {
	return nil, nil
}

func (s *Server) handleButton(w http.ResponseWriter, r *http.Request) {
	// Get button info from query params
	instanceID := chi.URLParam(r, "instanceId")
	deviceID := chi.URLParam(r, "deviceId")
	profileID := chi.URLParam(r, "profileId")
	pageID := chi.URLParam(r, "pageId")
	buttonID := chi.URLParam(r, "buttonId")

	// Get button buffer from NATS KV
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s.buffer",
		instanceID, deviceID, profileID, pageID, buttonID)

	entry, err := s.kv.Get(key)

	if err != nil {
		return
	}

	// Write buffer data to response
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(entry.Value())
}

func (s *Server) handleButtonPress(w http.ResponseWriter, r *http.Request) {
	nc, _ := natsconn.GetNATSConn()

	// Get button info from query params
	instanceID := chi.URLParam(r, "instanceId")
	deviceID := chi.URLParam(r, "deviceId")
	profileID := chi.URLParam(r, "profileId")
	pageID := chi.URLParam(r, "pageId")
	buttonID := chi.URLParam(r, "buttonId")

	var button, err = buttons.GetButton("instances." + instanceID + ".devices." + deviceID + ".profiles." + profileID + ".pages." + pageID + ".buttons." + buttonID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get button")
		return
	}

	buttonData, err := json.Marshal(button)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button")
		return
	}

	nc.Publish(button.UUID, buttonData)
}

func (s *Server) handleDeviceCardList(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handling device list request")
	instanceID := chi.URLParam(r, "instanceId")

	devices, err := devices.GetDevices(instanceID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get devices")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Interface("devices", devices).Msg("Found devices")
	partials.DeviceCardList(instanceID, devices).Render(r.Context(), w)
}

func (s *Server) handleInstanceCardList(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handling instance list request")

	instances, err := instance.GetInstances()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get instances")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Interface("instances", instances).Msg("Found instances")
	partials.InstanceCardList(instances).Render(r.Context(), w)
}

func (s *Server) Start() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Info().Msgf("Starting server on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}

func (s *Server) Close() {
}
