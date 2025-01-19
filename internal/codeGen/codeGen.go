package codegen

import (
	"bytes"
	codesandbox "codeGen/internal/codeSandbox"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

var workingMemory string

func promptCodeLlama(prompt string) string{
	data := RequestData{
		Model: "codellama",
		Prompt: fmt.Sprintf("%s Only respond with a full go file of code to resolve this question. Never respond with text that isn't part of the golang file used to resolve prompt. No explanations or comments. Do not add 'go' the the triple backticks of the response. Make sure all necessary imports are added to the code.", prompt),
		Stream: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	req, err := http.NewRequest("POST", "http://192.168.1.52:11434/api/generate", bytes.NewBuffer(jsonData))
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

func CodeGen() {
	codePrompt := "can you write a golang rest api that is used for user to sign-up and login to accounts?"

	codePromptResult := promptCodeLlama(codePrompt)
	
	codePromptCodeResult := strings.Split(codePromptResult, "\n")
	
	codesandbox.AddFileToSandbox("main.go", strings.Join(codePromptCodeResult[1 : len(codePromptCodeResult)-1], "\n"))

	depCheckResponse := checkForDeps(strings.Join(codePromptCodeResult[1 : len(codePromptCodeResult)-1], "\n"))

	fmt.Println(depCheckResponse)
	
	testPrompt := fmt.Sprintf("Use this golang code: %s, to create just the test to ensure the code to solve this works correctly.", codePromptResult)

	testPromptResult := promptCodeLlama(testPrompt)

	testPromptCodeResult := strings.Split(testPromptResult, "\n")

	codesandbox.AddFileToSandbox("main_test.go", strings.Join(testPromptCodeResult[1 : len(testPromptCodeResult)-1], "\n"))

	testGenCode()
}

func testGenCode() {
	cmd := exec.Command("go", "test")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error running tests: %v \nFixing this issues\n", err)
		fixTestIssues(string(output))
	}

	fmt.Println("test output:\n", string(output))
}

func fixTestIssues(errorOutput string) {
	failedTest := strings.Split(errorOutput, "\n")

	fileContents, err := os.ReadFile("./main.go")
	if err != nil {
		fmt.Printf("failed to read file, got error: %v", err)
	}

	testFileContents, err := os.ReadFile("./main_test.go")
	if err != nil {
		fmt.Printf("failed to read file, got error: %v", err)
	}

	failPrompt := fmt.Sprintf("This go file, %v, was tested with this test file, %v, and the result of the test from this test was, %v. Can you fix the issues that were found by this test?", fileContents, testFileContents, failedTest)

	failResult := promptCodeLlama(failPrompt)
	fmt.Println(failResult)
}

func checkForDeps(fileContents string) string {
	modFileContents, err := os.ReadFile("./go.mod")
	if err != nil {
		fmt.Printf("failed to read file, got error: %v", err)
	}

	data := RequestData{
		Model: "codellama",
		Prompt: fmt.Sprintf("Can you check this golang file, %v, and this go.mod file, %v, to check if any dependancies are missing. If any dependancies need to be installed only respond with the go get command the is used to install it, one go get command per line. Only respond with the command 'go get [missing dependency]' replacing [missing dependency] with the package that is missing.", fileContents, modFileContents),
		Stream: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	req, err := http.NewRequest("POST", "http://192.168.1.52:11434/api/generate", bytes.NewBuffer(jsonData))
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
