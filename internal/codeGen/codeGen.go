package codegen

import (
	"bytes"
	codesandbox "codeGen/internal/codeSandbox"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/client"
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

func promptCodeLlama(ctx context.Context, prompt string) string{
	data := RequestData{
		Model: "codellama",
		Prompt: fmt.Sprintf("%s Only respond with a full go file of code to resolve this question. Never respond with text that isn't part of the golang file used to resolve prompt. No explanations or comments. Do not add 'go' the the triple backticks of the response.", prompt),
		Stream: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://192.168.1.52:11434/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	var ollamaResponse OllamaResponse
	err = json.Unmarshal(body, &ollamaResponse)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	if resp.Status == "200 OK" {
		return ollamaResponse.Response
	}

	return ""
}

func CodeGen(ctx context.Context, cli *client.Client) {
	prompt := "can you create a golang function that adds two numbers?"

	testPrompt := fmt.Sprintf("Using the prompt: %s, create just the test to ensure the code to solve this works correctly.", prompt)
	
	testPromptResult := promptCodeLlama(ctx, testPrompt)
	codesandbox.AddFileToSandbox(ctx, cli, "./add_test.go", strings.Split(testPromptResult, "```")[1])
	fmt.Println(testPromptResult)
	codePrompt := fmt.Sprintf("Using this test file: %s, resolve this prompt %s", testPromptResult, prompt)
	codePromptResult := promptCodeLlama(ctx, codePrompt)
	codesandbox.AddFileToSandbox(ctx, cli, "./add.go", strings.Split(codePromptResult, "```")[1])
	fmt.Println(codePromptResult)
	return
}

func createFile(fileName string, content string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}
}
