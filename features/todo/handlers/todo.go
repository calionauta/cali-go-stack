package handlers

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/a-h/templ"
	"github.com/calionauta/cali-go-stack/config"
	"github.com/calionauta/cali-go-stack/features/todo"
	"github.com/calionauta/cali-go-stack/features/todo/components"
	dshelpers "github.com/calionauta/cali-go-stack/internal/datastar"
	"github.com/calionauta/cali-go-stack/internal/queue"
	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	sdk "github.com/starfederation/datastar-go/datastar"
)

type TodoHandler struct {
	app *pocketbase.PocketBase
	q   *queue.Queue
	cfg *config.Config
}

func New(app *pocketbase.PocketBase, q *queue.Queue, cfg *config.Config) *TodoHandler {
	return &TodoHandler{app: app, q: q, cfg: cfg}
}

func (h *TodoHandler) RegisterRoutes(se *core.ServeEvent) {
	se.Router.GET("/api/todos", h.handleList)
	se.Router.POST("/api/todos", h.handleCreate)
	se.Router.POST("/api/todos/{id}/toggle", h.handleToggle)
	se.Router.POST("/api/todos/completed/delete", h.handleClearCompleted)
	se.Router.POST("/api/todos/{id}/delete", h.handleDelete)
	se.Router.GET("/api/todos/stream", h.handleSSEStream)
}

func (h *TodoHandler) handleList(c *core.RequestEvent) error {
	filter := c.Request.URL.Query().Get("filter")
	todos, err := h.listTodos(filter)
	if err != nil {
		return c.String(500, "error listing todos")
	}

	sse := sdk.NewSSE(c.Response, c.Request)
	return dshelpers.MergeSignals(sse, todo.TodoSignals{
		Todos: todos, Filter: filter, ItemCount: len(todos),
	})
}

func (h *TodoHandler) handleCreate(c *core.RequestEvent) error {
	if err := c.Request.ParseForm(); err != nil {
		return c.String(400, "invalid form")
	}
	title := c.Request.FormValue("title")
	if title == "" {
		return c.String(400, "title required")
	}

	item := todo.Todo{
		ID:        uuid.New().String(),
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.saveTodo(item); err != nil {
		slog.Error("todo: save failed", "error", err)
		return c.String(500, "save failed")
	}

	data, _ := json.Marshal(map[string]string{
		"type": "todo_created", "todoID": item.ID, "title": item.Title,
	})
	if err := h.q.Enqueue(c.Request.Context(), data); err != nil {
		slog.Warn("todo: enqueue failed", "error", err)
	}

	todos, _ := h.listTodos("all")
	sse := sdk.NewSSE(c.Response, c.Request)
	return dshelpers.RenderAndPatch(sse, h.renderTodoList(todos))
}

func (h *TodoHandler) handleToggle(c *core.RequestEvent) error {
	rec, err := h.app.FindRecordById("todos", c.Request.PathValue("id"))
	if err != nil {
		return c.String(404, "not found")
	}
	rec.Set("completed", !rec.GetBool("completed"))
	if err := h.app.Save(rec); err != nil {
		slog.Error("todo: toggle save failed", "id", rec.Id, "error", err)
		return c.String(500, "toggle failed")
	}

	todos, _ := h.listTodos("all")
	sse := sdk.NewSSE(c.Response, c.Request)
	return dshelpers.RenderAndPatch(sse, h.renderTodoList(todos))
}

func (h *TodoHandler) handleDelete(c *core.RequestEvent) error {
	rec, err := h.app.FindRecordById("todos", c.Request.PathValue("id"))
	if err != nil {
		return c.String(404, "not found")
	}
	if err := h.app.Delete(rec); err != nil {
		slog.Error("todo: delete failed", "id", rec.Id, "error", err)
		return c.String(500, "delete failed")
	}

	todos, _ := h.listTodos("all")
	sse := sdk.NewSSE(c.Response, c.Request)
	return dshelpers.RenderAndPatch(sse, h.renderTodoList(todos))
}

func (h *TodoHandler) handleClearCompleted(c *core.RequestEvent) error {
	records, err := h.app.FindRecordsByFilter("todos", "completed=true", "", 0, 0)
	if err != nil {
		slog.Error("todo: find completed failed", "error", err)
		return c.String(500, "find failed")
	}
	for _, r := range records {
		if err := h.app.Delete(r); err != nil {
			slog.Warn("todo: clear-completed delete failed", "id", r.Id, "error", err)
		}
	}

	todos, _ := h.listTodos("all")
	sse := sdk.NewSSE(c.Response, c.Request)
	return dshelpers.RenderAndPatch(sse, h.renderTodoList(todos))
}

func (h *TodoHandler) handleSSEStream(c *core.RequestEvent) error {
	clientID := c.Request.URL.Query().Get("clientID")
	if clientID == "" {
		clientID = uuid.New().String()
	}

	sse := sdk.NewSSE(c.Response, c.Request)
	ch := make(chan []byte, 64)
	h.q.Hub().Register(clientID, ch)
	defer h.q.Hub().Unregister(clientID)

	todos, _ := h.listTodos("all")
	_ = dshelpers.MergeSignals(sse, todo.TodoSignals{
		Todos: todos, Filter: "all", ItemCount: len(todos),
	})

	for {
		select {
		case <-c.Request.Context().Done():
			return nil
		case msg := <-ch:
			_ = sse.MarshalAndPatchSignals(map[string]any{"toast": string(msg)})
		}
	}
}

// --- Repository ---

func (h *TodoHandler) listTodos(filter string) ([]todo.Todo, error) {
	var filterExpr string
	switch filter {
	case "active":
		filterExpr = "completed=false"
	case "completed":
		filterExpr = "completed=true"
	default:
		filterExpr = ""
	}
	records, err := h.app.FindRecordsByFilter("todos", filterExpr, "-created", 0, 0)
	if err != nil {
		return nil, err
	}
	res := make([]todo.Todo, len(records))
	for i, r := range records {
		res[i] = todoFromRecord(r)
	}
	return res, nil
}

func (h *TodoHandler) saveTodo(item todo.Todo) error {
	col, err := h.app.FindCollectionByNameOrId("todos")
	if err != nil {
		return err
	}
	rec := core.NewRecord(col)
	rec.Set("id", item.ID)
	rec.Set("title", item.Title)
	rec.Set("completed", item.Completed)
	return h.app.Save(rec)
}

func todoFromRecord(r *core.Record) todo.Todo {
	return todo.Todo{
		ID:        r.Id,
		Title:     r.GetString("title"),
		Completed: r.GetBool("completed"),
		CreatedAt: r.GetDateTime("created").Time(),
		UpdatedAt: r.GetDateTime("updated").Time(),
	}
}

func (h *TodoHandler) renderTodoList(todos []todo.Todo) templ.Component {
	return components.TodoList(todo.TodoSignals{
		Todos: todos, Filter: "all", ItemCount: len(todos),
	})
}
