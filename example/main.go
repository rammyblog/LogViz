package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rammyblog/logviz"
)

// Example handler
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	// time.Sleep(2.455 * time.Millisecond)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("Desewa, Aduke Ade"))
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
	// Register the Logger middleware
	mux.Handle("/", reqLogger.Logger(http.HandlerFunc(HelloHandler)))
	mux.Handle("/rammy", reqLogger.Logger(http.HandlerFunc(SecondHandler)))

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
