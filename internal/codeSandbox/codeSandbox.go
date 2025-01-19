package codesandbox

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateCodeSandbox() {
	err := os.Mkdir("codeSandbox", 0775)
	if err != nil {
		fmt.Printf("error creating directory with error: %v", err)
	}

	err = os.Chdir("./codeSandbox")
	if err != nil {
		fmt.Printf("error changing directory with error: %v", err)
	}

	cmd := exec.Command("go", "mod", "init", "codeSandbox")

	err = cmd.Run()
	if err != nil {
		fmt.Printf("error running commands with error: %v", err)
	}
}


func AddFileToSandbox(fileName string, content string) {
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