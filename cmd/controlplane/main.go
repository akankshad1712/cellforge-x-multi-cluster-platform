package main

import (
	"github.com/akankshad1712/cellforge/internal/platform"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	store := platform.NewStore()
	scheduler := platform.NewScheduler(store)
	api := platform.NewAPI(store, scheduler)
	srv := &http.Server{Addr: addr, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("CELLFORGE control plane listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
