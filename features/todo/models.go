package todo

import "time"

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created"`
	UpdatedAt time.Time `json:"updated"`
}

type TodoSignals struct {
	Todos      []Todo `json:"todos"`
	NewTitle   string `json:"newTitle"`
	Filter     string `json:"filter"` // "all", "active", "completed"
	EditingID  string `json:"editingId"`
	EditTitle  string `json:"editTitle"`
	Loading    bool   `json:"loading"`
	ItemCount  int    `json:"itemCount"`
}
