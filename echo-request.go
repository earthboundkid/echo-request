package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// subscribe to SIGINT signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s %q", r.URL, r.UserAgent())
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("Unexpected error: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "inline")
		_, err = w.Write(b)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	srv := &http.Server{Addr: ":" + port, Handler: http.DefaultServeMux}

	errc := make(chan error)
	go func() {
		log.Printf("Begin listening on port %s", port)
		// service connections
		errc <- srv.ListenAndServe()
	}()

	<-stopChan // wait for SIGINT
	log.Println("Shutting down server...")

	// shut down gracefully, but wait no longer than 5 seconds before halting
	ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()
	srv.Shutdown(ctx)

	select {
	case err := <-errc:
		log.Printf("Finished listening: %v\n", err)
	case <-ctx.Done():
		log.Println("Graceful shutdown timed out")
	}

	log.Println("Server stopped")
}
