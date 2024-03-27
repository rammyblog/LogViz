package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rammyblog/logviz"
)

// Example handler
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	time.Sleep(2 * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}

func SecondHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("second handler"))
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	reqLogger, err := logviz.Init("postgres", logviz.DbConfig{
		DbUser:     "postgres",
		DbPassword: "password",
		DbHost:     "localhost",
		DbName:     "logviz",
		DbPort:     "5429",
	})

	if err != nil {
		fmt.Println(err)
	}
	reqLogger.Serve(":5009")

	mux.Handle("/", reqLogger.Logger(http.HandlerFunc(HelloHandler)))
	mux.Handle("/second", reqLogger.Logger(http.HandlerFunc(SecondHandler)))

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
