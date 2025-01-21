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
	"regexp"
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
		Model: "deepseek-r1:8b",
		Prompt: fmt.Sprintf("Here is the working context of the pervious conversation: %s.\n\n%s Only respond with a full go file of code to resolve this question. Never respond with text that isn't part of the golang file used to resolve prompt. No explanations or comments. Do not add 'go' the the triple backticks of the response. Make sure all necessary imports are added to the code.", workingMemory, prompt),
		Stream: false,
	}

	workingMemory = workingMemory + "\npervious prompt: " + data.Prompt

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
	codePrompt := "can you write a simple go rest api?"

	codePromptResult := promptCodeLlama(codePrompt)

	re := regexp.MustCompile(`(?s)<think>.*?</think>`)

	removeThink := re.ReplaceAllString(codePromptResult, "")
	
	codePromptCodeResult := strings.Split(removeThink, "\n")
	
	workingMemory = workingMemory + "\npervious response: " + strings.Join(codePromptCodeResult[1 : len(codePromptCodeResult)-1], "\n")

	codesandbox.AddFileToSandbox("main.go", strings.Join(codePromptCodeResult[1 : len(codePromptCodeResult)-1], "\n"))
	
	testPrompt := fmt.Sprintf("Use this golang code: %s, to create just the test to ensure the code to solve this works correctly.", codePromptResult)

	testPromptResult := promptCodeLlama(testPrompt)

	removeCodeThink := re.ReplaceAllString(testPromptResult, "")
	

	testPromptCodeResult := strings.Split(removeCodeThink, "\n")

	workingMemory = workingMemory + "\npervious response: " + strings.Join(testPromptCodeResult[1 : len(testPromptCodeResult)-1], "\n")

	codesandbox.AddFileToSandbox("main_test.go", strings.Join(testPromptCodeResult[1 : len(testPromptCodeResult)-1], "\n"))

	testGenCode()
}

func testGenCode() {
	cmd := exec.Command("go", "test")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error running tests: %v\n", err)
		fmt.Printf("test output: %v\nRerunning codeGen\n", string(output))
		err := os.Remove("main.go")
		if err != nil {
			fmt.Printf("error deleting main.go with error: %v", err)
		}
		err = os.Remove("main_test.go")
		if err != nil {
			fmt.Printf("error deleting main.go with error: %v", err)
		}
		workingMemory = workingMemory + "\nThe pervious two prompts and responses resulted in an error: " + string(output) + ". Use this broken code as context to return working code."
		CodeGen()
	} else {
		fmt.Println("test output:\n", string(output))
	}
}
