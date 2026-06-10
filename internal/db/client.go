package db

import (
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
)

func InitPocketBase() *pocketbase.PocketBase {
	app := pocketbase.New()

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
		os.Args = append(os.Args, "--http=0.0.0.0:8090")
	}

	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("PocketBase failed to start: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	return app
}
