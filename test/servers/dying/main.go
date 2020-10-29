package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func Fatalf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(-1)
}

func main() {
	mux := http.NewServeMux()
	go func() {
		time.Sleep(3 * time.Second)
		Fatalf("Let's die!\n")
	}()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `
# HELP answer_to_everything Answer To Everything
# TYPE answer_to_everything gauge
answer_to_everything{scope="universe",env="prod"} 42
`)
	})
	srv := &http.Server{
		Addr:    "localhost:9999",
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		Fatalf("Failed to run web server: %v", err)
	}
}
