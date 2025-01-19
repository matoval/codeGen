package main

import (
	codegen "codeGen/internal/codeGen"
	codesandbox "codeGen/internal/codeSandbox"
)

func main() {
	codesandbox.CreateCodeSandbox()
	codegen.CodeGen()
}
