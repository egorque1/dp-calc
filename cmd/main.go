package main

import (
	"log"
	"net/http"
	"os"

	"dp-calc/internal/handles"
)

func main() {
	port := os.Getenv("PORT") // Render задаёт PORT через переменные окружения
	if port == "" {
		port = "8080" // локально по умолчанию
	}
	http.HandleFunc("/", handles.HandleIndex)
	http.HandleFunc("/params", handles.HandleParams)
	http.HandleFunc("/compute", handles.HandleCompute)
	http.HandleFunc("/download", handles.HandleDownload)

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
