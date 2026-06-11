package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"grow/internal/models"
)

type NotionSyncResult struct {
	SyncedAt string `json:"synced_at"`
	Updated  int    `json:"updated"`
	Created  int    `json:"created"`
	Total    int    `json:"total"`
}

type NotionSearchRequest struct {
	Filter NotionSearchFilter `json:"filter"`
}

type NotionSearchFilter struct {
	Property string `json:"property"`
	Title    NotionTitleFilter `json:"title"`
}

type NotionTitleFilter struct {
	Equals string `json:"equals"`
}

type NotionSearchResponse struct {
	Results []NotionPage `json:"results"`
}

type NotionPage struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
}

func SyncToNotion(db *sql.DB, apiKey, databaseID string) (*NotionSyncResult, error) {
	abilities, err := models.ListAbilities(db)
	if err != nil {
		return nil, fmt.Errorf("list abilities: %w", err)
	}

	ApplyDecayToMany(abilities)

	result := &NotionSyncResult{
		SyncedAt: time.Now().Format("2006-01-02 15:04:05"),
		Total:    len(abilities),
	}

	for _, a := range abilities {
		// Check if page exists (search by title)
		pageID, err := findNotionPage(apiKey, databaseID, a.Name)
		if err != nil {
			log.Printf("Notion: search page for '%s': %v", a.Name, err)
			continue
		}

		if pageID != "" {
			if err := updateNotionPage(apiKey, pageID, &a); err != nil {
				log.Printf("Notion: update page for '%s': %v", a.Name, err)
				continue
			}
			result.Updated++
		} else {
			if err := createNotionPage(apiKey, databaseID, &a); err != nil {
				log.Printf("Notion: create page for '%s': %v", a.Name, err)
				continue
			}
			result.Created++
		}
	}

	log.Printf("Notion sync: %d updated, %d created, %d total", result.Updated, result.Created, result.Total)
	return result, nil
}

func findNotionPage(apiKey, databaseID, title string) (string, error) {
	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Name",
			"title": map[string]interface{}{
				"equals": title,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.notion.com/v1/databases/"+databaseID+"/query", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("notion search error %d: %s", resp.StatusCode, string(respBody))
	}

	var result NotionSearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Results) > 0 {
		return result.Results[0].ID, nil
	}
	return "", nil
}

func updateNotionPage(apiKey, pageID string, a *models.Ability) error {
	lastActivity := "N/A"
	if a.LastActivityAt != nil {
		lastActivity = a.LastActivityAt.Format("2006-01-02")
	}

	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"Current Value":   map[string]interface{}{"number": a.CurrentValue},
			"Effective Value": map[string]interface{}{"number": a.EffectiveValue},
			"Base Value":      map[string]interface{}{"number": a.BaseValue},
			"Decay Rate":      map[string]interface{}{"number": a.DecayRate},
			"Last Activity":   map[string]interface{}{"date": map[string]interface{}{"start": lastActivity}},
		},
	}

	return notionAPI(apiKey, "PATCH", "https://api.notion.com/v1/pages/"+pageID, payload)
}

func createNotionPage(apiKey, databaseID string, a *models.Ability) error {
	lastActivity := "N/A"
	if a.LastActivityAt != nil {
		lastActivity = a.LastActivityAt.Format("2006-01-02")
	}

	payload := map[string]interface{}{
		"parent": map[string]interface{}{"database_id": databaseID},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]interface{}{"content": a.Name}},
				},
			},
			"Current Value":   map[string]interface{}{"number": a.CurrentValue},
			"Effective Value": map[string]interface{}{"number": a.EffectiveValue},
			"Base Value":      map[string]interface{}{"number": a.BaseValue},
			"Decay Rate":      map[string]interface{}{"number": a.DecayRate},
			"Last Activity":   map[string]interface{}{"date": map[string]interface{}{"start": lastActivity}},
		},
	}

	return notionAPI(apiKey, "POST", "https://api.notion.com/v1/pages", payload)
}

func notionAPI(apiKey, method, url string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notion api error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
