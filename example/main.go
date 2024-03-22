package main

import (
	"log"
	"net/http"
)

// Example handler
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register the Logger middleware
	// mux.Handle("/", logviz.Logger(http.HandlerFunc(HelloHandler)))

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
