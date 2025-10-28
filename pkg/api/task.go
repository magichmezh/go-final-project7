package api

import (
	"encoding/json"
	"fmt"
	"go1f/pkg/db"
	"net/http"
	"strings"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Задача не найдена"})
		return
	}

	writeJson(w, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": "Ошибка десериализации JSON"})
		return
	}
	if strings.TrimSpace(task.ID) == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор задачи"})
		return
	}
	if strings.TrimSpace(task.Title) == "" {
		writeJson(w, map[string]string{"error": "Не указано название задачи"})
		return
	}
	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": "Некорректная дата или правило повторения"})
		return
	}
	err := db.UpdateTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка обновления задачи: %v", err)})
		return
	}
	writeJson(w, map[string]string{})
}
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	err := db.DeleteTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": "Ошибка удаления задачи: " + err.Error()})
		return
	}
	writeJson(w, map[string]string{})
}
