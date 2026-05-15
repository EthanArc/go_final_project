package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EthanArc/go_final_project/pkg/db"
)

const readerLimit = 1048576

// registering URL paths
func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", nextDayHandler)
	mux.HandleFunc("/api/task", taskHandler)
	mux.HandleFunc("/api/tasks", tasksHandler)
	mux.HandleFunc("/api/task/done", taskDoneHandler)
}

func taskHandler(resWri http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		addTaskHandler(resWri, req)
	case http.MethodGet:
		getTaskHandler(resWri, req)
	case http.MethodPut:
		updateTaskHandler(resWri, req)
	case http.MethodDelete:
		deleteTaskHandler(resWri, req)
	default:
		http.Error(resWri, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func addTaskHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost { //Drops any traffic that is not HTTP
		sendJSONResponse(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	//if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") { // Checking state
	//	sendJSONResponse(resWri, http.StatusUnsupportedMediaType, map[string]string{"error": "Expected JSON content"})
	//	return
	//}
	contentType := strings.ToLower(req.Header.Get("Content-Type"))

	if !strings.Contains(contentType, "application/json") {
		sendJSONResponse(resWri, http.StatusUnsupportedMediaType, map[string]string{"error": "Expected JSON content"})
		return
	}

	req.Body = http.MaxBytesReader(resWri, req.Body, readerLimit) // 1MB limit
	defer req.Body.Close()

	var task db.Task
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		sendJSONResponse(resWri, http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		log.Printf("JSON decode error: %v", err)
		return
	}

	if task.Title == "" {
		sendJSONResponse(resWri, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	if err := checkDate(&task); err != nil {
		sendJSONResponse(resWri, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid date: %v", err)})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		sendJSONResponse(resWri, http.StatusInternalServerError, map[string]string{"error": "Failed to add task"})
		log.Printf("DB error: %v", err)
		return
	}

	sendJSONResponse(resWri, http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func nextDayHandler(resWri http.ResponseWriter, req *http.Request) {

	//Validate - only get allowed
	if req.Method != http.MethodGet {
		http.Error(resWri, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Paesing query params
	query := req.URL.Query()
	nowStr := query.Get("now")
	date := query.Get("date")
	repeat := query.Get("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now() // Today bt default
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(resWri, "Invalid now parameter format", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(resWri, err.Error(), http.StatusBadRequest)
		return
	}

	//Send response
	resWri.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := resWri.Write([]byte(nextDate)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
