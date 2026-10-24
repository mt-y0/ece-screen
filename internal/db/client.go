package db

import (
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// startupTimeout bounds how long InitPocketBase waits for the server to come up.
const startupTimeout = 30 * time.Second

// InitPocketBase starts the embedded PocketBase server and blocks until it is
// bootstrapped, the required collections exist and the HTTP server is about to
// listen — so callers can query the database immediately.
//
// First run (no pb_data/): the schema is created with example records
// (see schema.go) and PocketBase prints — and opens in the browser — the
// one-time URL for registering the first superuser account.
func InitPocketBase() *pocketbase.PocketBase {
	app := pocketbase.New()

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
		os.Args = append(os.Args, "--http=0.0.0.0:8090")
	}

	ready := make(chan struct{})
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := EnsureSchema(e.App); err != nil {
			return err // aborts serve; surfaced via Start()'s error below
		}
		close(ready)
		return e.Next()
	})

	done := make(chan error, 1)
	go func() { done <- app.Start() }()

	select {
	case <-ready:
	case err := <-done:
		if err != nil {
			log.Fatalf("PocketBase failed to start: %v", err)
		}
		// A non-serve command (e.g. `superuser upsert`) finished; nothing to display.
		os.Exit(0)
	case <-time.After(startupTimeout):
		log.Fatalf("PocketBase did not become ready within %s", startupTimeout)
	}

	// Keep reporting fatal server errors after startup.
	go func() {
		if err := <-done; err != nil {
			log.Fatalf("PocketBase stopped: %v", err)
		}
	}()

	return app
}
