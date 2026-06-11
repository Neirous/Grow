package handlers

import (
	"net/http"

	"grow/internal/db"
	"grow/internal/models"
	"grow/internal/service"
)

func RegisterDashboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard", Dashboard)
}

type DashboardResponse struct {
	Abilities  []models.Ability `json:"abilities"`
	Activities []models.Activity `json:"activities"`
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	abilities, err := models.ListAbilities(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if abilities == nil {
		abilities = []models.Ability{}
	}
	service.ApplyDecayToMany(abilities)

	activities, err := models.ListActivities(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if activities == nil {
		activities = []models.Activity{}
	}

	writeJSON(w, http.StatusOK, DashboardResponse{
		Abilities:  abilities,
		Activities: activities,
	})
}
