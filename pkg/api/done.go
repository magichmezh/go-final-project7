package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go1f/pkg/db"
	"log"
)

func doneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		respondJSONError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		respondJSONError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	log.Printf("[doneHandler] id=%s date=%s repeat=%q", task.ID, task.Date, task.Repeat)

	if strings.TrimSpace(task.Repeat) == "" {
		err = db.DeleteTask(id)
		if err != nil {
			respondJSONError(w, http.StatusInternalServerError, "Ошибка удаления задачи: "+err.Error())
			log.Printf("[doneHandler] delete id=%s error=%v", id, err)
			return
		}
		log.Printf("[doneHandler] deleted id=%s", id)
	} else {
		d, errp := time.ParseInLocation(dateFormat, task.Date, time.Now().Location())
		if errp != nil {
			respondJSONError(w, http.StatusBadRequest, "Ошибка разбора даты задачи: "+errp.Error())
			log.Printf("[doneHandler] parse date error id=%s date=%s err=%v", id, task.Date, errp)
			return
		}
		next, err := NextDate(d, task.Date, task.Repeat)
		if err != nil {
			respondJSONError(w, http.StatusBadRequest, "Ошибка вычисления следующей даты: "+err.Error())
			log.Printf("[doneHandler] nextdate error id=%s err=%v", id, err)
			return
		}
		err = db.UpdateDate(next, id)
		if err != nil {
			respondJSONError(w, http.StatusInternalServerError, "Ошибка обновления даты: "+err.Error())
			log.Printf("[doneHandler] update id=%s next=%s error=%v", id, next, err)
			return
		}
		log.Printf("[doneHandler] updated id=%s next=%s", id, next)
	}

	respondJSON(w, http.StatusOK, map[string]string{})
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func respondJSONError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}
