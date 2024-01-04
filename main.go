package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Tested")
        fmt.Fprintf(w, "Hello, this is a test endpoint!")
    })

    fmt.Println("Server starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}