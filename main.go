package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ParaResult struct {
	Results []RunnerOutput
}

type RunnerOutput struct {
	Command       string
	Raw           string
	Json          map[string]interface{}
	ExecutionTime string
}

func main() {
	commands := readFromStdin()
	results := handler(commands)
	render(results)
}

func render(results ParaResult) {
	output, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}

	fmt.Printf("%s\n", output)
}

func readFromStdin() []string {
	commands := []string{}

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Failed getting stats from stdin: %s", err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// no piped data
		return commands
	}

	// https://stackoverflow.com/a/12369689
	inputStream := bufio.NewScanner(os.Stdin)
	for inputStream.Scan() {
		cmd := strings.TrimSpace(inputStream.Text())

		if len(cmd) > 0 {
			commands = append(commands, cmd)
		}
	}

	return commands
}

func handler(commands []string) ParaResult {
	n := len(commands)
	runnerOutput := make(chan RunnerOutput, n)
	outputs := []RunnerOutput{}

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go runner(commands[i], &wg, runnerOutput)
	}

	wg.Wait()

	for i := 0; i < n; i++ {
		outputs = append(outputs, <-runnerOutput)
	}

	return ParaResult{Results: outputs}
}

func runner(cmd string, wg *sync.WaitGroup, output chan RunnerOutput) {
	defer wg.Done()

	start := time.Now()
	out, _ := exec.Command("sh", "-c", cmd).CombinedOutput()
	elapsed := time.Since(start)

	var rawJson map[string]interface{}
	json.Unmarshal(out, &rawJson)

	output <- RunnerOutput{
		Command:       cmd,
		Raw:           string(out[:]),
		Json:          rawJson,
		ExecutionTime: fmt.Sprintf("%s", elapsed),
	}
}
