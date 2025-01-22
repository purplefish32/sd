package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sd/cmd/web/views/partials"
	"sd/pkg/buttons"
	"sd/pkg/devices"
	"sd/pkg/instance"
	"sd/pkg/natsconn"
	"sd/pkg/profiles"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func HandleButton(w http.ResponseWriter, r *http.Request) {
	_, kv := natsconn.GetNATSConn()

	// Get button info from query params
	instanceID := chi.URLParam(r, "instanceId")
	deviceID := chi.URLParam(r, "deviceId")
	profileID := chi.URLParam(r, "profileId")
	pageID := chi.URLParam(r, "pageId")
	buttonID := chi.URLParam(r, "buttonId")

	// Get button buffer from NATS KV
	key := fmt.Sprintf("instances.%s.devices.%s.profiles.%s.pages.%s.buttons.%s.buffer",
		instanceID, deviceID, profileID, pageID, buttonID)

	entry, err := kv.Get(key)

	if err != nil {
		return
	}

	// Write buffer data to response
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(entry.Value())
}

func HandleButtonPress(w http.ResponseWriter, r *http.Request) {
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

func HandleDeviceCardList(w http.ResponseWriter, r *http.Request) {
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

func HandleInstanceCardList(w http.ResponseWriter, r *http.Request) {
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

func HandleProfileAddDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")

		component := partials.ProfileAddDialog(instanceID, deviceID)
		component.Render(r.Context(), w)
	}
}

func HandleProfileCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	instanceID := r.FormValue("instanceId")
	deviceID := r.FormValue("deviceId")
	name := r.FormValue("name")

	log.Info().Str("instanceId", instanceID).Str("deviceId", deviceID).Str("name", name).Msg("Creating profile")

	_, err = profiles.CreateProfile(instanceID, deviceID, name)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create profile")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profiles, err := profiles.GetProfiles(instanceID, deviceID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partials.ProfileCardList(instanceID, deviceID, profiles).Render(r.Context(), w)
}

func HandleProfileDeleteDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		profile, err := profiles.GetProfile(instanceID, deviceID, profileID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to get profile")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		component := partials.ProfileDeleteDialog(instanceID, deviceID, *profile)
		component.Render(r.Context(), w)
	}
}

func HandleProfileDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		err := profiles.DeleteProfile(instanceID, deviceID, profileID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete profile")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		profiles, err := profiles.GetProfiles(instanceID, deviceID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to get profiles after deletion")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		partials.ProfileCardList(instanceID, deviceID, profiles).Render(r.Context(), w)
	}
}
