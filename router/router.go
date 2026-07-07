package router

import (
	"net/http"

	"github.com/calionauta/cali-go-stack/config"
	"github.com/calionauta/cali-go-stack/features/todo/handlers"
	"github.com/calionauta/cali-go-stack/internal/queue"
	"github.com/calionauta/cali-go-stack/web/resources"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Init registers custom routes on PocketBase's serve event.
// Call before pb.Start().
func Init(app *pocketbase.PocketBase, q *queue.Queue, cfg *config.Config) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/health", func(c *core.RequestEvent) error {
			return c.String(200, "ok")
		})

		se.Router.GET("/static/*", func(c *core.RequestEvent) error {
			fs := http.StripPrefix("/static/", http.FileServer(resources.StaticFS()))
			fs.ServeHTTP(c.Response, c.Request)
			return nil
		})

		// Register example feature: Todo MVC
		h := handlers.New(app, q, cfg)
		h.RegisterRoutes(se)

		return se.Next()
	})
}
