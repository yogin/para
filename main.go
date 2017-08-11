package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

type RunnerOutput struct {
	Command string
	Raw     string
	Json    map[string]interface{}
}

func main() {
	// https://coderwall.com/p/zyxyeg/golang-having-fun-with-os-stdin-and-shell-pipes
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Failed getting stats from stdin: %s", err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		// no piped data
		return
	}

	inputStream := bufio.NewScanner(os.Stdin)
	commands := []string{}

	// https://stackoverflow.com/a/12369689
	for inputStream.Scan() {
		inputCommand := inputStream.Text()
		commands = append(commands, inputCommand)
	}

	n := len(commands)
	runnerOutput := make(chan RunnerOutput, n)
	results := []RunnerOutput{}

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go runner(commands[i], &wg, runnerOutput)
	}

	wg.Wait()

	for i := 0; i < n; i++ {
		results = append(results, <-runnerOutput)
	}

	data, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}

	fmt.Printf("%s\n", data)
}

func runner(cmd string, wg *sync.WaitGroup, output chan RunnerOutput) {
	defer wg.Done()

	out, _ := exec.Command("sh", "-c", cmd).CombinedOutput()

	var rawJson map[string]interface{}
	json.Unmarshal(out, &rawJson)

	output <- RunnerOutput{
		Command: cmd,
		Raw:     string(out[:]),
		Json:    rawJson,
	}
}
