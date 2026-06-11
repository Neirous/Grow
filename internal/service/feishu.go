package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"grow/internal/models"
)

type FeishuCardMessage struct {
	MsgType string       `json:"msg_type"`
	Card    FeishuCard   `json:"card"`
}

type FeishuCard struct {
	Header  FeishuCardHeader   `json:"header"`
	Elements []FeishuCardElement `json:"elements"`
}

type FeishuCardHeader struct {
	Title FeishuCardText `json:"title"`
}

type FeishuCardText struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type FeishuCardElement struct {
	Tag  string          `json:"tag"`
	Text *FeishuCardText `json:"text,omitempty"`
	Note *struct {
		Elements []FeishuCardText `json:"elements"`
	} `json:"note,omitempty"`
}

func SendFeishuReminder(db *sql.DB, webhookURL string) error {
	abilities, err := models.ListAbilities(db)
	if err != nil {
		return fmt.Errorf("list abilities: %w", err)
	}

	ApplyDecayToMany(abilities)

	now := time.Now()
	var lines []string
	needPractice := false

	for _, a := range abilities {
		if a.DaysSinceLastActivity > 0.5 { // more than 12 hours
			needPractice = true
			pct := (a.CurrentValue - a.EffectiveValue) / a.CurrentValue * 100
			lines = append(lines, fmt.Sprintf("• %s（上次：%.0f天前，有效值已衰减 %.1f%%）",
				a.Name, a.DaysSinceLastActivity, pct))
		}
	}

	if !needPractice {
		lines = append(lines, "🎉 所有能力都在活跃状态，继续保持！")
	}

	msg := FeishuCardMessage{
		MsgType: "interactive",
		Card: FeishuCard{
			Header: FeishuCardHeader{
				Title: FeishuCardText{Content: "grow 每日提醒", Tag: "plain_text"},
			},
			Elements: []FeishuCardElement{
				{Tag: "div", Text: &FeishuCardText{
					Content: "今日还有以下能力待练习：\n" + strings.Join(lines, "\n"),
					Tag: "lark_md",
				}},
				{Tag: "note", Note: &struct {
					Elements []FeishuCardText `json:"elements"`
				}{
					Elements: []FeishuCardText{
						{Content: fmt.Sprintf("📊 %s 发送 | 打开 grow 开始练习 💪", now.Format("2006-01-02 15:04")), Tag: "plain_text"},
					},
				}},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}

	log.Println("Feishu reminder sent successfully")
	return nil
}
