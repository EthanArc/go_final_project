package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// Global database connection instance
//var db *sql.DB

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// ErrDBNotInit is returned when operations run before database connection setup
var ErrDBNotInit = errors.New("database connection is not initialized")

func AddTask(task *Task) (int64, error) {
	if db == nil {
		return 0, ErrDBNotInit
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?);`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return id, nil
}

func Tasks(limit int) ([]*Task, error) {
	if db == nil {
		return nil, ErrDBNotInit
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?;`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	result := make([]*Task, 0, limit)

	for rows.Next() {
		task := &Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		result = append(result, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return result, nil
}

func GetTask(id string) (*Task, error) {
	if db == nil {
		return nil, ErrDBNotInit
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?;`
	var task Task
	err := db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch task: %w", err)
	}
	return &task, nil
}

func UpdateTask(task *Task) error {

	if db == nil {
		return ErrDBNotInit
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?;`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return errors.New("incorrect id for updating task")
	}
	return nil
}

func DeleteTask(id string) error {
	if db == nil {
		return ErrDBNotInit
	}

	query := `DELETE FROM scheduler WHERE id = ?;`
	res, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return errors.New("task not found")
	}
	return nil
}

func UpdateDate(next, id string) error {
	if db == nil {
		return ErrDBNotInit
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?;`
	res, err := db.Exec(query, next, id)
	if err != nil {
		return fmt.Errorf("failed to update date: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if count == 0 {
		return errors.New("incorrect id for updating date")
	}
	return nil
}
