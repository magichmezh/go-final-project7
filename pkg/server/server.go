package server

import (
	"encoding/json"
	"fmt"
	"go1f/pkg/api"
	"go1f/pkg/db"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
)

func Run() error {
	// Choose DB file. Tests expect ../scheduler.db by default; allow override
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "../scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		return fmt.Errorf("error initializing db: %w", err)
	}

	port := 7540
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	api.Init()

	http.HandleFunc("/tasks", tasksHandler)

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	fmt.Printf("Server started at http://localhost:%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), recoverMiddleware(http.DefaultServeMux))
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprint(rec)})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// Унифицированная отправка JSON ошибки
func respondJSONError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasks(w)
	case http.MethodPost:
		createTask(w, r)
	default:
		respondJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func getTasks(w http.ResponseWriter) {
	tasks, err := db.Tasks(50)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string][]*db.Task{"tasks": tasks})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respondJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if t.Date == "" || t.Title == "" {
		respondJSONError(w, http.StatusBadRequest, "Missing required fields: date or title")
		return
	}
	res, err := db.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	lastID, err := res.LastInsertId()
	if err == nil {
		t.ID = fmt.Sprint(lastID)
	}
	respondJSON(w, http.StatusCreated, t)
}
