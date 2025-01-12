package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sd/cmd/web/views/pages"
	"sd/cmd/web/views/pages/components"
	"sd/pkg/natsconn"
)

type Server struct {
	router *chi.Mux
	log    zerolog.Logger
	nc     *nats.Conn
	kv     nats.KeyValue
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
	// Serve static files
	fileServer := http.FileServer(http.Dir("cmd/web/static"))
	s.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Routes
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pages.Home().Render(r.Context(), w)
	})

	s.router.Get("/devices", func(w http.ResponseWriter, r *http.Request) {
		pages.Devices().Render(r.Context(), w)
	})

	s.router.Get("/profiles", func(w http.ResponseWriter, r *http.Request) {
		pages.Profiles().Render(r.Context(), w)
	})

	s.router.Get("/settings", func(w http.ResponseWriter, r *http.Request) {
		pages.Settings().Render(r.Context(), w)
	})

	// HTMX Routes
	s.router.Get("/devices/list", s.handleDeviceList)
	s.router.Get("/devices/{id}", s.handleDeviceConfig)
	s.router.Get("/instances/list", s.handleInstanceList)
	s.router.Get("/instances/{id}/devices", s.handleInstanceDevices)

	// API Routes
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/devices", s.handleGetDevices)
	})
}

func (s *Server) getDevices() ([]components.Device, error) {
	keys, err := s.kv.Keys()
	if err != nil {
		return nil, err
	}

	devices := make([]components.Device, 0)
	seen := make(map[string]bool)

	for _, key := range keys {
		if strings.Contains(key, "devices") {
			parts := strings.Split(key, ".")
			if len(parts) < 4 {
				continue
			}

			deviceID := parts[3]
			if seen[deviceID] {
				continue // Skip duplicates
			}

			devices = append(devices, components.Device{
				ID:       deviceID,
				Instance: parts[1],
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

func (s *Server) handleDeviceConfig(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	// TODO: Load device config from NATS
	components.DeviceConfig(deviceID).Render(r.Context(), w)
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
	keys, err := s.kv.Keys()
	if err != nil {
		return nil, err
	}

	devices := make([]components.Device, 0)
	seen := make(map[string]bool)

	prefix := "instances." + instanceID + ".devices."
	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			parts := strings.Split(key, ".")
			if len(parts) < 4 {
				continue
			}

			deviceID := parts[3]
			if seen[deviceID] {
				continue
			}

			// Get device details from KV store
			entry, err := s.kv.Get(key)
			if err != nil {
				continue
			}

			var deviceInfo struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(entry.Value(), &deviceInfo); err != nil {
				continue
			}

			devices = append(devices, components.Device{
				ID:       deviceID,
				Instance: instanceID,
				Type:     deviceInfo.Type,
			})
			seen[deviceID] = true
		}
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
