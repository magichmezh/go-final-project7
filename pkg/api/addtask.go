package api

import (
	"encoding/json"
	"log"
	"go1f/pkg/db"
	"net/http"
	"strings"
	"time"
)

func checkDate(task *db.Task) error {
	now := time.Now()
	if strings.TrimSpace(task.Date) == "" {
		task.Date = now.Format("20060102")
	}
	t, err := time.ParseInLocation("20060102", task.Date, now.Location())
	if err != nil {
		return err
	}
	if strings.TrimSpace(task.Repeat) != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		if nowDate.After(tDate) {
			log.Printf("[checkDate] nowDate=%v tDate=%v next=%s", nowDate, tDate, next)
			task.Date = next
		}
	} else if now.After(t) {
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		task.Date = nowDate.Format("20060102")
	}
	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "Ошибка десериализации JSON"})
		return
	}
	if strings.TrimSpace(task.Title) == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}
	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": "Некорректная дата или правило повторения: " + err.Error()})
		return
	}
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "Ошибка при добавлении задачи: " + err.Error()})
		return
	}
	writeJson(w, map[string]int64{"id": id})
}
