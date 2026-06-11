package handlers

import (
	"net/http"
	"time"

	"grow/internal/db"
)

type StreakResponse struct {
	CurrentStreak int              `json:"current_streak"`
	LongestStreak int              `json:"longest_streak"`
	Heatmap       []HeatmapDay     `json:"heatmap"`
}

type HeatmapDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func RegisterStreakRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/streaks", GetStreaks)
}

func GetStreaks(w http.ResponseWriter, r *http.Request) {
	// Get daily activity counts for the past year
	rows, err := db.DB.Query(`
		SELECT DATE(completed_at) as day, COUNT(*) as cnt
		FROM activity_logs
		WHERE completed_at >= datetime('now', '-1 year')
		GROUP BY day
		ORDER BY day ASC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	dailyMap := make(map[string]int)
	for rows.Next() {
		var day string
		var cnt int
		if err := rows.Scan(&day, &cnt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dailyMap[day] = cnt
	}

	// Build heatmap for past 365 days
	var heatmap []HeatmapDay
	now := time.Now()
	currentStreak := 0
	longestStreak := 0
	tempStreak := 0

	for i := 364; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		count := dailyMap[dateStr]
		heatmap = append(heatmap, HeatmapDay{Date: dateStr, Count: count})

		if count > 0 {
			tempStreak++
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
		} else {
			tempStreak = 0
		}
	}

	// Calculate current streak (from today backwards)
	today := now.Format("2006-01-02")
	todayCount := dailyMap[today]
	// If no activity today, check from yesterday
	checkDate := now
	if todayCount == 0 {
		checkDate = now.AddDate(0, 0, -1)
	}
	for {
		dateStr := checkDate.Format("2006-01-02")
		if dailyMap[dateStr] > 0 {
			currentStreak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	// Apply the overall longest streak
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	writeJSON(w, http.StatusOK, StreakResponse{
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		Heatmap:       heatmap,
	})
}
