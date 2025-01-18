package main

import (
	codegen "codeGen/internal/codeGen"
	codesandbox "codeGen/internal/codeSandbox"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received termination signal (Ctrl+C or SIGTERM)")
		cancel()
	}()

	cli := codesandbox.CreateCodeSandbox(ctx)
	codegen.CodeGen(ctx, cli)

	select {
	case <-ctx.Done():
		codesandbox.RemoveContainer(ctx, cli)
	}

}
