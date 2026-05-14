package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	//	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/EthanArc/go_final_project/pkg/db"
)

// Seting template
const dateFormat = "20060102"

func sendJSONResponse(resWri http.ResponseWriter, status int, data any) {
	resWri.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resWri.WriteHeader(status)
	if err := json.NewEncoder(resWri).Encode(data); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func addTaskHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost { //Drops any traffic that is not HTTP
		sendJSONResponse(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	if !strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") { // Checking state
		sendJSONResponse(resWri, http.StatusUnsupportedMediaType, map[string]string{"error": "Expected JSON content"})
		return
	}

	req.Body = http.MaxBytesReader(resWri, req.Body, 1048576) // 1MB limit
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

func checkDate(task *db.Task) error {
	nowTime := time.Now()

	now := nowTime.Truncate(24 * time.Hour)

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	timPar, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return err
	}

	if timPar.Before(now) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format(dateFormat)
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
			task.Date = next
		}
	}
	return nil
}
