package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "github.com/joho/godotenv"
    "os"
    "strings"
    "io"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

type ScriptRequest struct {
    Endpoint string `json:"endpoint"`
    Method   string `json:"method"`
    Vus string `json:"vus"`
    DurationInSecond string `json:"durationInSecond"`
}

type ChatCompletionResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int    `json:"created"`
    Model   string `json:"model"`
    Choices []Choice `json:"choices"`
}

type Choice struct {
    Index   int    `json:"index"`
    Message Message `json:"message"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Response struct {
    Message string `json:"message"`
}


func main() {
    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Tested")
        fmt.Fprintf(w, "Hello, this is a test endpoint!")
    })

    http.HandleFunc("/generate", generateScriptHandler)

    fmt.Println("Server starting on port 8080...")

    // test := ScriptRequest{
    //     Endpoint : "test",
    //     Method : "GET",
    //     Vus : 5,
    //     DurationInSecond : 5,
    // }

    // GenerateScript(test)
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCors(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}


func generateScriptHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("123123")

	enableCors(&w)
    if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
        return
    }
    var req ScriptRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }


    script, err := GenerateScript(req)
	if err != nil{
		fmt.Fprintln(w, "Error generating script")
	}

	w.Header().Set("Content-Type", "text/html")

	// response := Response{
    //     Message: script,
    // }

    fmt.Printf("Generated k6 script for testing %s method on %s", req.Method, req.Endpoint)

	// // Encode the response as JSON
	// if err := json.NewEncoder(w).Encode(response); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	w.Write(script)
}


func GenerateScript(rq ScriptRequest) ([]byte, error){

	err := godotenv.Load()
	if err != nil {
        log.Println(err)
		log.Fatal("Error loading .env file")
	}

	secretKey := os.Getenv("APIKEY")

	fmt.Println("SECRET KEY IS " + secretKey)

	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(
		`{
			"model": "gpt-3.5-turbo",
			"messages": [
				{
				"role": "system",
				"content": "Greet the user first. You are a assistant that generates k6 load testing script file."
				},
				{
				"role": "assistant",
				"content": "Hello! I am a assistant that can help you generate k6 load testing script file. How can I assist you today?"
				},
				{
				"role": "user",
				"content": "Create a k6 script for load testing for the endpoint %v. This should be a %v request. Simulate %v users over a period of %v seconds."
				}
			],
			"temperature": 1,
			"max_tokens": 256,
			"top_p": 1,
			"frequency_penalty": 0,
			"presence_penalty": 0
			}`, rq.Endpoint, rq.Method, rq.Vus, rq.DurationInSecond,
	))

	fmt.Println(payload)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization",fmt.Sprintf("Bearer %v", secretKey))
	req.Header.Add("Cookie", "__cf_bm=JUD4mqGL8D5L8euOSvs2IfwtVK0FriIpBu4zwyW8Drw-1704602516-1-AQyGW9C96nQlU3uPDTkX1YW+0N+D6447rO+Lrx9gG06ot167LYqr6YMJq05PXa004wvSpbZWpgortn/Pc+A2VSY=; _cfuvid=nKMl6J8V0OsI5w7ls5J4536jjIZXeKQW6hEcCUHiWsc-1704600670478-0-604800000")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}

    var response ChatCompletionResponse
    err = json.Unmarshal([]byte(string(body)), &response)
    if err != nil {
        log.Fatalf("Error occurred during unmarshalling: %s", err)
    }

    fmt.Println("Content:", response.Choices[0].Message.Content)

	md := []byte([]byte(response.Choices[0].Message.Content))
	html := mdToHTML(md)
	fmt.Println(string(body))

	return html, nil

}