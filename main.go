package main

import (
	"fmt"
	"sync"
	"encoding/json"
	"log"
	"os/exec"
)

type RunnerOutput struct {
	Command string
	Output string
}

func main() {
	fmt.Println("vim-go")

	n := 5

	runnerOutput := make(chan RunnerOutput, n)
	results := []RunnerOutput{}

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go runner("ls -al", &wg, runnerOutput)
	}

	for i := 0; i < n; i++ {
		results = append(results, <- runnerOutput)
	}

	wg.Wait()

	data, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func runner(cmd string, wg *sync.WaitGroup, output chan RunnerOutput) {
	defer wg.Done()
	fmt.Println(fmt.Sprintf("runner: %s", cmd))
	out, _ := exec.Command("sh", "-c", cmd).Output()

	output <- RunnerOutput{
		Command: cmd,
		Output: string(out[:]),
	}
}

