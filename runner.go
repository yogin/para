package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	Command       string
	Raw           string
	Json          map[string]interface{}
	ExecutionTime string
}

func NewRunner(cmd string) *Runner {
	return &Runner{Command: cmd}
}

func NewRunnersFromBuffer(buffer io.Reader) []*Runner {
	var runners []*Runner

	stream := bufio.NewScanner(buffer)
	for stream.Scan() {
		cmd := strings.TrimSpace(stream.Text())

		if len(cmd) > 0 {
			runners = append(runners, NewRunner(cmd))
		}
	}

	return runners
}

func (r *Runner) Run() {
	start := time.Now()
	out, _ := exec.Command("sh", "-c", r.Command).CombinedOutput()
	r.ExecutionTime = time.Since(start).String()
	r.Raw = string(out[:])
	json.Unmarshal(out, &r.Json)
}
