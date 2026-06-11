package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"grow/internal/db"
	"grow/internal/models"
)

type CreateActivityRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Effects     []models.ActivityEffect `json:"effects"`
}

func RegisterActivityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/activities", ListActivities)
	mux.HandleFunc("POST /api/activities", CreateActivity)
	mux.HandleFunc("GET /api/activities/{id}", GetActivity)
	mux.HandleFunc("PUT /api/activities/{id}", UpdateActivity)
	mux.HandleFunc("DELETE /api/activities/{id}", DeleteActivity)
}

func ListActivities(w http.ResponseWriter, r *http.Request) {
	activities, err := models.ListActivities(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if activities == nil {
		activities = []models.Activity{}
	}
	writeJSON(w, http.StatusOK, activities)
}

func CreateActivity(w http.ResponseWriter, r *http.Request) {
	var req CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	a := &models.Activity{
		Name:        req.Name,
		Description: req.Description,
		Effects:     req.Effects,
	}
	if a.Effects == nil {
		a.Effects = []models.ActivityEffect{}
	}

	if err := models.CreateActivity(db.DB, a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload to get effects with ability names
	created, err := models.GetActivity(db.DB, a.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func GetActivity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	a, err := models.GetActivity(db.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func UpdateActivity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	_, err = models.GetActivity(db.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "activity not found")
		return
	}

	var req CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	a := &models.Activity{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Effects:     req.Effects,
	}
	if a.Effects == nil {
		a.Effects = []models.ActivityEffect{}
	}

	if err := models.UpdateActivity(db.DB, a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, _ := models.GetActivity(db.DB, id)
	writeJSON(w, http.StatusOK, updated)
}

func DeleteActivity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := models.DeleteActivity(db.DB, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
