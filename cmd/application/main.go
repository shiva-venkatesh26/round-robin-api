package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8081"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		log.Printf("Got a hit on an instance of Application API running on port %s", port)
		w.Write([]byte(fmt.Sprintf(`{"message":"Echo from Application API on port %s"}`, port)))
	})

	log.Printf("Application API running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
