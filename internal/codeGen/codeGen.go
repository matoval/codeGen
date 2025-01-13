package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type RequestData struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool `json:"stream"`
}

type OllamaResponse struct {
	Model string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response string `json:"response"`
	Done bool `json:"done"`
}

func promptCodeLlama(prompt string) string{
	data := RequestData{
		Model: "codellama",
		Prompt: fmt.Sprintf("%s Only respond with a full go file of code to resolve this question. Never respond with text that isn't part of the golang file used to resolve prompt.", prompt),
		Stream: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("http://192.168.1.52:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var ollamaResponse OllamaResponse
	err = json.Unmarshal(body, &ollamaResponse)
	if err != nil {
		panic(err)
	}

	if resp.Status == "200 OK" {
		return ollamaResponse.Response
	}

	return ""
}

func CodeGen() {
	prompt := "can you create a golang function that adds two numbers?"

	testPrompt := fmt.Sprintf("Using the prompt: %s, create just the test to ensure the code to solve this works correctly.", prompt)
	
	testPromptResult := promptCodeLlama(testPrompt)
	createFile( "./add_test.go", strings.Split(testPromptResult, "```")[1])
	fmt.Println(testPromptResult)
	codePrompt := fmt.Sprintf("Using this test file: %s, resolve this prompt %s", testPromptResult, prompt)
	codePromptResult := promptCodeLlama(codePrompt)
	createFile("./add.go", strings.Split(codePromptResult, "```")[1])
	fmt.Println(codePromptResult)
}

func createFile(fileName string, content string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}
}
