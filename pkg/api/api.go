package api

import (
	"net/http"
	"time"
)

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
	resWri.Write([]byte(nextDate))
}
