package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"sd/cmd/web/handlers"
	"sd/cmd/web/views/partials"
	"sd/pkg/natsconn"
	"sd/pkg/store"
	"sd/pkg/types"
)

type Server struct {
	router *chi.Mux
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
	}

	// Setup routes
	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// Routes
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		instances := store.GetInstances()
		partials.HomePage(instances).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}", func(w http.ResponseWriter, r *http.Request) {
		instanceID := chi.URLParam(r, "instanceID")

		instances := store.GetInstances()
		devices := store.GetDevices(instanceID)
		instance := store.GetInstance(instanceID)
		partials.InstancePage(instance, instances, devices).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}/device/{deviceID}", func(w http.ResponseWriter, r *http.Request) {
		instances := store.GetInstances()
		instanceID := chi.URLParam(r, "instanceID")
		deviceID := chi.URLParam(r, "deviceID")

		device := store.GetDevice(instanceID, deviceID)

		devices := store.GetDevices(instanceID)
		profiles := store.GetProfiles(instanceID, device)
		pages := store.GetPages(instanceID, device, profiles[0].ID) // TODO: Fix this
		instance := store.GetInstance(instanceID)
		partials.DevicePage(instances, devices, profiles, pages, instance, device).Render(r.Context(), w)
	})

	s.router.Get("/instance/{instanceID}/device/{deviceID}/profile/{profileID}/page/{pageID}", func(w http.ResponseWriter, r *http.Request) {
		instanceID := chi.URLParam(r, "instanceID")
		deviceID := chi.URLParam(r, "deviceID")
		profileID := chi.URLParam(r, "profileID")
		pageID := chi.URLParam(r, "pageID")

		instances := store.GetInstances()
		devices := store.GetDevices(instanceID)
		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)
		profile := store.GetProfile(instanceID, device, profileID)
		page := store.GetPage(instanceID, deviceID, profileID, pageID)
		profiles := store.GetProfiles(instanceID, device)
		pages := store.GetPages(instanceID, device, profileID)
		partials.ProfilePage(instances, devices, profiles, pages, instance, device, profile, page).Render(r.Context(), w)

		// Set the current page and profile in the nats  KV store using the store package
		// store.SetCurrentPage(instanceID, deviceID, profileID, pageID)
		// store.SetCurrentProfile(instanceID, deviceID, profileID)

	})

	// HTMX Routes
	s.router.Get("/partials/instance-card-list", handlers.HandleInstanceCardList)
	s.router.Get("/partials/{instanceId}/device-card-list", handlers.HandleDeviceCardList)
	s.router.Get("/partials/button/{instanceId}/{deviceId}/{profileId}/{pageId}/{buttonId}", handlers.HandleButton)
	s.router.Post("/partials/button/{instanceId}/{deviceId}/{profileId}/{pageId}/{buttonId}", handlers.HandleButtonPress)
	s.router.Get("/partials/profile/add", handlers.HandleProfileAddDialog())
	s.router.Get("/partials/close-dialog", func(w http.ResponseWriter, r *http.Request) {
		// Return empty response to remove the dialog
		w.Write([]byte(""))
	})
	s.router.Get("/partials/profile/delete-dialog", handlers.HandleProfileDeleteDialog())
	s.router.Get("/partials/page/delete-dialog", handlers.HandlePageDeleteDialog())

	// Add SSE endpoint for device updates
	s.router.Get("/stream/instance/{instanceId}/devices", func(w http.ResponseWriter, r *http.Request) {
		_, kv := natsconn.GetNATSConn()
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
		watcher, err := kv.Watch("instances." + instanceID + ".devices.*")

		if err != nil {
			log.Error().Err(err).Msg("Failed to create KV watcher")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer watcher.Stop()

		// 4. Handle client disconnection
		go func() {
			<-r.Context().Done()
			done <- true
		}()

		// 5. Main event loop
		for {
			select {
			case <-done:
				return
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				// Re-fetch the updated devices list instead of reusing the stale one
				updatedDevices := store.GetDevices(instanceID)
				// Also, optionally re-fetch the instance if it could have changed:
				updatedInstance := store.GetInstance(instanceID)

				if err := s.sendDeviceList(w, r.Context(), updatedInstance, updatedDevices); err != nil {
					if err != context.Canceled {
						log.Error().Err(err).Msg("Failed to send device list")
					}
					return
				}
			}
		}
	})

	// Add SSE endpoint for profile updates
	s.router.Get("/stream/instance/{instanceId}/device/{deviceId}/profiles", func(w http.ResponseWriter, r *http.Request) {
		_, kv := natsconn.GetNATSConn()
		instanceID := chi.URLParam(r, "instanceId")
		deviceID := chi.URLParam(r, "deviceId")

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
		watcher, err := kv.Watch("instances." + instanceID + ".devices." + deviceID + ".profiles.*")

		if err != nil {
			log.Error().Err(err).Msg("Failed to create KV watcher")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		defer watcher.Stop()

		// 4. Handle client disconnection
		go func() {
			<-r.Context().Done()
			done <- true
		}()

		// 5. Main event loop
		for {
			select {
			case <-done:
				return
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				// Re-fetch the updated data on each watcher update
				updatedDevice := store.GetDevice(instanceID, deviceID)
				updatedInstance := store.GetInstance(instanceID)
				updatedProfiles := store.GetProfiles(instanceID, updatedDevice)

				if err := s.sendProfileList(w, r.Context(), updatedInstance, updatedDevice, updatedProfiles); err != nil {
					if err != context.Canceled {
						log.Error().Err(err).Msg("Failed to send profile list")
					}
					return
				}
			}
		}
	})

	// REST API routes.

	s.router.Post("/api/profile/create", handlers.HandleProfileCreate)
	s.router.Delete("/api/profile/delete", handlers.HandleProfileDelete())

	s.router.Post("/api/page/create", handlers.HandlePageCreate)
	s.router.Delete("/api/page", handlers.HandlePageDelete())
}

func (s *Server) sendDeviceList(w http.ResponseWriter, ctx context.Context, instance types.Instance, devices []types.Device) error {
	var buf bytes.Buffer

	if err := partials.DeviceCardList(instance, devices).Render(ctx, &buf); err != nil {
		return err
	}

	fmt.Fprintf(w, "event: DeviceCardListUpdate\n")
	fmt.Fprintf(w, "data: %s\n\n", buf.String())

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

func (s *Server) sendProfileList(w http.ResponseWriter, ctx context.Context, instance types.Instance, device *types.Device, profiles []types.Profile) error {
	var buf bytes.Buffer

	if err := partials.ProfileCardList(instance, device, profiles).Render(ctx, &buf); err != nil {
		return err
	}

	fmt.Fprintf(w, "event: ProfileCardListUpdate\n")
	fmt.Fprintf(w, "data: %s\n\n", buf.String())

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
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

// Add this method to server.Server
func (s *Server) Router() *chi.Mux {
	return s.router
}
