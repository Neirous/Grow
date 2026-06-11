package handlers

import (
	"encoding/json"
	"net/http"

	"grow/internal/db"
	"grow/internal/models"
)

func RegisterSettingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", GetSettings)
	mux.HandleFunc("PUT /api/settings", UpdateSettings)
}

func GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := models.GetAllSettings(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if settings == nil {
		settings = make(map[string]string)
	}
	writeJSON(w, http.StatusOK, settings)
}

func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := models.SetAllSettings(db.DB, settings); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
