package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
)

type ScriptRequest struct {
    Endpoint string `json:"endpoint"`
    Method   string `json:"method"`
}


func main() {
    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Tested")
        fmt.Fprintf(w, "Hello, this is a test endpoint!")
    })

    http.HandleFunc("/generate", generateScriptHandler)

    fmt.Println("Server starting on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}


func generateScriptHandler(w http.ResponseWriter, r *http.Request) {
    var req ScriptRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Generate script based on JSON body
    script := fmt.Sprintf("Generated k6 script for testing %s method on %s", req.Method, req.Endpoint)

    fmt.Fprintln(w, script)
}