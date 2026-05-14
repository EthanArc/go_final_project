package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EthanArc/go_final_project/pkg/db"
)

type TaskResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// Helper to standardise responses and headers
func sendJSRes(resWri http.ResponseWriter, statusCode int, data interface{}) {
	resWri.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resWri.WriteHeader(statusCode)
	_ = json.NewEncoder(resWri).Encode(data)
}

// Helper to extract and validate ID URL parameters
func parseIDParam(req *http.Request) (string, error) {
	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		return "", errors.New("Id is null")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return "", errors.New("Invalid id")
	}
	return idStr, nil
}

func tasksHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		sendJSRes(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Expected GET"})
		return
	}

	tasks, err := db.Tasks(50)
	if err != nil {
		sendJSRes(resWri, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if tasks == nil {
		tasks = []*db.Task{}
	}
	sendJSRes(resWri, http.StatusOK, TaskResp{Tasks: tasks})
}

func getTaskHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		sendJSRes(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Expected GET"})
		return
	}

	idStr, err := parseIDParam(req)
	if err != nil {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	task, err := db.GetTask(idStr)
	if err != nil {
		sendJSRes(resWri, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	sendJSRes(resWri, http.StatusOK, task)
}

func updateTaskHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut && req.Method != http.MethodPost {
		sendJSRes(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Expected POST or PUT"})
		return
	}

	var task db.Task
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if task.Title == "" {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": "Title is required"})
		return
	}

	if checkDate(&task) != nil {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": "Format date wrong"})
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		sendJSRes(resWri, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	sendJSRes(resWri, http.StatusOK, map[string]interface{}{})
}

func deleteTaskHandler(resWri http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		sendJSRes(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Expected DELETE"})
		return
	}

	idStr, err := parseIDParam(req)
	if err != nil {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := db.DeleteTask(idStr); err != nil {
		sendJSRes(resWri, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sendJSRes(resWri, http.StatusOK, map[string]interface{}{})
}

func taskDoneHandler(resWri http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSRes(resWri, http.StatusMethodNotAllowed, map[string]string{"error": "Expected POST"})
		return
	}

	idStr, err := parseIDParam(r)
	if err != nil {
		sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	task, err := db.GetTask(idStr)
	if err != nil {
		sendJSRes(resWri, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if task.Repeat != "" {
		next, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			sendJSRes(resWri, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := db.UpdateDate(next, idStr); err != nil {
			sendJSRes(resWri, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	} else {
		if err := db.DeleteTask(idStr); err != nil {
			sendJSRes(resWri, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	sendJSRes(resWri, http.StatusOK, map[string]interface{}{})
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("Empty repeat")
	}

	ruleRepeat := strings.Split(repeat, " ")
	if len(ruleRepeat) == 0 || len(ruleRepeat) > 2 {
		return "", fmt.Errorf("Expecting 1 or 2 arguments, received %d", len(ruleRepeat))
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", err
	}

	switch ruleRepeat[0] {
	case "d":
		if len(ruleRepeat) < 2 {
			return "", errors.New("Expecting 2 arguments, received 1")
		}
		interval, err := strconv.Atoi(ruleRepeat[1])
		if err != nil {
			return "", err
		}
		if interval <= 0 || interval > 400 {
			return "", fmt.Errorf("Max d 400")
		}

		// Calculate total days behind 'now'
		if date.Before(now) {
			daysDiff := int(now.Sub(date).Hours() / 24)
			// Fast-forward past 'now' using math
			intervalsNeeded := (daysDiff / interval) + 1
			date = date.AddDate(0, 0, intervalsNeeded*interval)

			// Adjust for potential shifts
			if !date.After(now) {
				date = date.AddDate(0, 0, interval)
			}
		} else {
			date = date.AddDate(0, 0, interval)
		}

	case "y":
		if len(ruleRepeat) != 1 {
			return "", errors.New("Expecting 1 argument for y")
		}

		// Fast-forward years using math
		if date.Before(now) {
			yearsDiff := now.Year() - date.Year()
			date = date.AddDate(yearsDiff, 0, 0)

			// If still not after 'now', add one more year
			if !date.After(now) {
				date = date.AddDate(1, 0, 0)
			}
		} else {
			date = date.AddDate(1, 0, 0)
		}

	default:
		return "", errors.New("Expecting d num or y")
	}

	return date.Format(dateFormat), nil
}
