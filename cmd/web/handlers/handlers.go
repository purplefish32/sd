package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sd/cmd/web/views/partials"
	"sd/pkg/natsconn"
	"sd/pkg/store"

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

	device := store.GetDevice(instanceID, deviceID)

	_, err = store.CreateProfile(instanceID, device, name)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create profile")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleProfileDeleteDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)
		profile := store.GetProfile(instanceID, device, profileID)

		component := partials.ProfileDeleteDialog(instance, device, profile)
		component.Render(r.Context(), w)
	}
}

func HandleProfileDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")

		device := store.GetDevice(instanceID, deviceID)

		err := store.DeleteProfile(instanceID, device, profileID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete profile")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		instance := store.GetInstance(instanceID)
		profiles := store.GetProfiles(instanceID, device)

		partials.ProfileCardList(instance, device, profiles).Render(r.Context(), w)
	}
}

func HandlePageCreate(w http.ResponseWriter, r *http.Request) {
	instanceID := r.URL.Query().Get("instanceId")
	deviceID := r.URL.Query().Get("deviceId")
	profileID := r.URL.Query().Get("profileId")

	device := store.GetDevice(instanceID, deviceID)

	page, err := store.CreatePage(instanceID, device, profileID)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set as current page
	// store.SetCurrentPage(instanceID, deviceID, profileID, page.ID)

	// Create blank buttons for the page
	// for i := 0; i < 32; i++ {
	// 	store.CreateButton(instanceID, device, profileID, page.ID, strconv.Itoa(i+1))
	// }

	// Get updated data
	instance := store.GetInstance(instanceID)
	device = store.GetDevice(instanceID, deviceID)
	profile := store.GetProfile(instanceID, device, profileID)
	pages := store.GetPages(instanceID, device, profileID)

	// Re-render the entire profile page
	partials.ProfilePage(
		store.GetInstances(),
		store.GetDevices(instanceID),
		store.GetProfiles(instanceID, device),
		pages,
		instance,
		device,
		profile,
		page,
	).Render(r.Context(), w)
}

func HandlePageDeleteDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")
		pageID := r.URL.Query().Get("pageId")

		instance := store.GetInstance(instanceID)
		device := store.GetDevice(instanceID, deviceID)
		profile := store.GetProfile(instanceID, device, profileID)
		page := store.GetPage(instanceID, deviceID, profileID, pageID)

		component := partials.PageDeleteDialog(instance, device, profile, page)
		component.Render(r.Context(), w)
	}
}

func HandlePageDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instanceID := r.URL.Query().Get("instanceId")
		deviceID := r.URL.Query().Get("deviceId")
		profileID := r.URL.Query().Get("profileId")
		pageID := r.URL.Query().Get("pageId")

		device := store.GetDevice(instanceID, deviceID)

		err := store.DeletePage(instanceID, device, profileID, pageID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to delete page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		store.DeletePage(instanceID, device, profileID, pageID)

		var profile = store.GetProfile(instanceID, device, profileID)

		log.Info().Interface("profile", profile).Msg("Profile")

		var previousPageID = profile.Pages[len(profile.Pages)-1].ID

		log.Info().Str("previousPageID", previousPageID).Msg("Previous Page ID")
		// store.SetCurrentPage(instanceID, deviceID, profileID, previousPageID)

		w.Header().Add("Hx-Redirect", "/instance/"+instanceID+"/device/"+deviceID+"/profile/"+profileID+"/page/"+previousPageID)
	}
}
