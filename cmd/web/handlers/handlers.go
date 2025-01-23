package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sd/cmd/web/views/partials"
	"sd/pkg/natsconn"
	"sd/pkg/store"
	"strconv"

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
		log.Error().Err(err).Str("key", key).Msg("Failed to get button buffer")
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

	var button, err = store.GetButton("instances." + instanceID + ".devices." + deviceID + ".profiles." + profileID + ".pages." + pageID + ".buttons." + buttonID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get button")
		return
	}

	buttonData, err := json.Marshal(button)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal button")
		return
	}

	nc.Publish(button.ID, buttonData)
}

func HandleDeviceCardList(w http.ResponseWriter, r *http.Request) {
	instanceID := chi.URLParam(r, "instanceId")
	instance := store.GetInstance(instanceID)
	devices := store.GetDevices(instanceID)
	partials.DeviceCardList(instance, devices).Render(r.Context(), w)
}

func HandleInstanceCardList(w http.ResponseWriter, r *http.Request) {
	instances := store.GetInstances()
	partials.InstanceCardList(instances).Render(r.Context(), w)
}

func HandleProfileAddDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")

		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)

		component := partials.ProfileAddDialog(instance, device)
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

	instance := store.GetInstance(instanceID)
	device := store.GetDevice(instanceID, deviceID)

	profile, err := store.CreateProfile(instanceID, deviceID, name)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create profile")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page, err := store.CreatePage(instanceID, deviceID, profile.ID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := 0; i < 32; i++ { // TODO: Make this configurable
		store.CreateButton(instanceID, deviceID, profile.ID, page.ID, strconv.Itoa(i))
	}

	profiles := store.GetProfiles(instanceID, deviceID)
	partials.ProfileCardList(instance, device, profiles).Render(r.Context(), w)
}

func HandleProfileDeleteDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)
		profile := store.GetProfile(instanceID, deviceID, profileID)

		component := partials.ProfileDeleteDialog(instance, device, profile)
		component.Render(r.Context(), w)
	}
}

func HandleProfileDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		err := store.DeleteProfile(instanceID, deviceID, profileID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete profile")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)
		profiles := store.GetProfiles(instanceID, deviceID)

		partials.ProfileCardList(instance, device, profiles).Render(r.Context(), w)
	}
}
