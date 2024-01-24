package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/joho/godotenv"
)

func listFilesInDir(dirPath string) ([]string, error) {
    var fileNames []string

    files, err := ioutil.ReadDir(dirPath)
    if err != nil {
        return nil, err
    }

    for _, file := range files {
        if !file.IsDir() { // or file.Mode().IsRegular() to include only regular files
            fileNames = append(fileNames, file.Name())
        }
    }

    return fileNames, nil
}

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
	http.HandleFunc("/savescript", saveCodeToFileHandler)

	http.HandleFunc("/run-k6/", runK6Handler) // Note the trailing slash

	http.HandleFunc("/list-files", func(w http.ResponseWriter, r *http.Request) {
        fileNames, err := listFilesInDir("k6-scripts")
        if err != nil {
            http.Error(w, "Unable to list files", http.StatusInternalServerError)
            return
        }

        tmpl, err := template.ParseFiles("listFiles.html")
        if err != nil {
            http.Error(w, "Unable to load template", http.StatusInternalServerError)
            return
        }

        tmpl.Execute(w, fileNames)
    })


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

func runK6Handler(w http.ResponseWriter, r *http.Request) {
    // Extract filename from URL path
    filename := path.Base(r.URL.Path)
    if filename == "" || filename == "run-k6" {
        http.Error(w, "Filename is required", http.StatusBadRequest)
        return
    }

    // Validate the filename
    if !isValidFilename(filename) {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }

    // Execute k6 command
	 // Create a buffer to capture output
	var stdout, stderr bytes.Buffer
    cmd := exec.Command("k6", "run", "./k6-scripts/"+filename)
	cmd.Stdout = &stdout // Capture standard output
    cmd.Stderr = &stderr // Capture standard error
    // Execute k6 command
    err := cmd.Run()
    outStr, errStr := stdout.String(), stderr.String()

    if err != nil {
        log.Printf("Error running k6 command: %v, stderr: %v", err, errStr)
        http.Error(w, "Error executing k6 command", http.StatusInternalServerError)
        return
    }

    // Logging the output
    log.Printf("k6 command output: %s", outStr)

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("k6 command executed successfully\n" + outStr))
}

func isValidFilename(filename string) bool {
    // Implement filename validation logic here
    // For example, ensure filename has a .js extension, no invalid characters, etc.
    return strings.HasSuffix(filename, ".js") // Simple example; enhance as needed
}

func saveCodeToFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("123123")

	enableCors(&w)
    if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
        return
    }
	fmt.Println(r.Body)
    bodyBytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(w, "can't read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

	err = saveCodeToFile(bodyBytes, "./k6-scripts/")

	if err != nil{
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
	}

    // Do something with the body string
    log.Printf("Body: %s", string(bodyBytes))

	w.WriteHeader(http.StatusOK) // Explicitly setting status to 200 OK
}

// saveCodeToFile saves a given code block (byte slice) to a specified JavaScript file path with a unique timestamp
func saveCodeToFile(codeBlock []byte, filePath string) error {
    // Append a timestamp to the filename to ensure uniqueness
    timestamp := time.Now().Format("20060102150405") // YYYYMMDDHHMMSS format
    filePath = filePath + "_" + timestamp + ".js"

    // Write data to file
    err := ioutil.WriteFile(filePath, codeBlock, 0644) // 0644 for readable and writable file
    if err != nil {
        return err
    }
    return nil
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
				"content": "Greet the user first. You are a assistant that generates k6 load testing script file. Do not provide instruction for the user on how to execute the test."
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