package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"grow/internal/db"
	"grow/internal/models"
	"grow/internal/service"
)

type CreateAbilityRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	BaseValue   float64 `json:"base_value"`
	CurrentValue float64 `json:"current_value"`
	GrowthRate  float64 `json:"growth_rate"`
	DecayRate   float64 `json:"decay_rate"`
}

func RegisterAbilityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/abilities", ListAbilities)
	mux.HandleFunc("POST /api/abilities", CreateAbility)
	mux.HandleFunc("GET /api/abilities/{id}", GetAbility)
	mux.HandleFunc("PUT /api/abilities/{id}", UpdateAbility)
	mux.HandleFunc("DELETE /api/abilities/{id}", DeleteAbility)
	mux.HandleFunc("GET /api/abilities/{id}/history", GetAbilityHistory)
}

func ListAbilities(w http.ResponseWriter, r *http.Request) {
	abilities, err := models.ListAbilities(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if abilities == nil {
		abilities = []models.Ability{}
	}
	service.ApplyDecayToMany(abilities)
	writeJSON(w, http.StatusOK, abilities)
}

func CreateAbility(w http.ResponseWriter, r *http.Request) {
	var req CreateAbilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.CurrentValue == 0 {
		req.CurrentValue = req.BaseValue
	}
	if req.GrowthRate == 0 {
		req.GrowthRate = 1.0
	}
	if req.DecayRate == 0 {
		req.DecayRate = 0.5
	}

	a := &models.Ability{
		Name:        req.Name,
		Description: req.Description,
		BaseValue:   req.BaseValue,
		CurrentValue: req.CurrentValue,
		GrowthRate:  req.GrowthRate,
		DecayRate:   req.DecayRate,
	}
	if err := models.CreateAbility(db.DB, a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

func GetAbility(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	a, err := models.GetAbility(db.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ability not found")
		return
	}

	service.ApplyDecay(a)
	writeJSON(w, http.StatusOK, a)
}

func UpdateAbility(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	existing, err := models.GetAbility(db.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "ability not found")
		return
	}

	var req CreateAbilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	if req.BaseValue != 0 {
		existing.BaseValue = req.BaseValue
	}
	if req.CurrentValue != 0 {
		existing.CurrentValue = req.CurrentValue
	}
	if req.GrowthRate != 0 {
		existing.GrowthRate = req.GrowthRate
	}
	if req.DecayRate != 0 {
		existing.DecayRate = req.DecayRate
	}

	if err := models.UpdateAbility(db.DB, existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func DeleteAbility(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := models.DeleteAbility(db.DB, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func GetAbilityHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := service.GetAbilityWithHistory(db.DB, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
