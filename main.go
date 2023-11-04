package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx, canc := context.WithCancel(context.Background())

	r := chi.NewRouter()

	// A realistic base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/debug", middleware.Profiler())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Let's allocate loads of memory
		data := make([]byte, 1024*1024)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Allocated %d bytes", len(data))))

	})

	httpServer := &http.Server{
		Addr:        ":8080",
		Handler:     r,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	httpServer.RegisterOnShutdown(canc)

	go func() {
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	<-signalChan
	log.Println("Shutting down...")

	go func() {
		<-signalChan
		log.Fatal("os.Kill - terminating...\n")
	}()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
		defer os.Exit(1)
		return
	} else {
		log.Println("Server gracefully stopped")
	}

	defer os.Exit(0)
	return
}
