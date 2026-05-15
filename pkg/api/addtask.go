package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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
