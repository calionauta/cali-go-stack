package morpheus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// marshalSignals serializes the todo.Signals struct as a JSON string for
// the data-signals attribute. Same logic as features/todo/components/marshalSignals.
func marshalSignals(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// safeJSStringMorpheus returns s as a JSON string literal with surrounding
// double quotes, suitable for embedding inside a Datastar event expression.
func safeJSStringMorpheus(s string) string {
	return strconv.Quote(s)
}

// timeAgo returns a human-readable relative time, matching the format
// used in features/todo/components/todo_item.templ.
func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "agora"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min"
		}
		return fmt.Sprintf("%d min", m)
	default:
		return t.Format("02/01")
	}
}
