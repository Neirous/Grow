package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"grow/internal/db"
	"grow/internal/models"
)

type CreateGoalRequest struct {
	AbilityID   int64  `json:"ability_id"`
	TargetValue float64 `json:"target_value"`
	Deadline    string `json:"deadline"`
}

func RegisterGoalRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/goals", ListGoals)
	mux.HandleFunc("POST /api/goals", CreateGoal)
	mux.HandleFunc("PUT /api/goals/{id}", UpdateGoal)
	mux.HandleFunc("DELETE /api/goals/{id}", DeleteGoal)
}

func ListGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := models.ListGoals(db.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if goals == nil {
		goals = []models.Goal{}
	}
	writeJSON(w, http.StatusOK, goals)
}

func CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.AbilityID == 0 || req.TargetValue <= 0 {
		writeError(w, http.StatusBadRequest, "ability_id and target_value are required")
		return
	}

	g := &models.Goal{
		AbilityID:   req.AbilityID,
		TargetValue: req.TargetValue,
	}
	if req.Deadline != "" {
		t, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid deadline format, use YYYY-MM-DD")
			return
		}
		g.Deadline = &t
	}

	if err := models.CreateGoal(db.DB, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, g)
}

func UpdateGoal(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	g := &models.Goal{
		ID:          id,
		AbilityID:   req.AbilityID,
		TargetValue: req.TargetValue,
		IsAchieved:  false,
	}
	if req.Deadline != "" {
		t, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid deadline format")
			return
		}
		g.Deadline = &t
	}

	if err := models.UpdateGoal(db.DB, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func DeleteGoal(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := models.DeleteGoal(db.DB, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
