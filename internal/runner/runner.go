package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

// Payload is sent as JSON body for webhook actions.
type Payload struct {
	Port  int    `json:"port"`
	Host  string `json:"host"`
	State string `json:"state"`
	Time  string `json:"time"`
}

// Runner executes actions (webhook or shell command) on port state changes.
type Runner struct {
	client *http.Client
}

// New returns a Runner with a default HTTP client.
func New() *Runner {
	return &Runner{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// RunWebhook sends a POST request to the given URL with a JSON payload.
func (r *Runner) RunWebhook(url string, p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("runner: marshal payload: %w", err)
	}

	resp, err := r.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("runner: webhook POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("runner: webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

// RunCommand executes a shell command string via sh -c.
func (r *Runner) RunCommand(command string, p Payload) error {
	env := []string{
		fmt.Sprintf("PW_PORT=%d", p.Port),
		fmt.Sprintf("PW_HOST=%s", p.Host),
		fmt.Sprintf("PW_STATE=%s", p.State),
		fmt.Sprintf("PW_TIME=%s", p.Time),
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(cmd.Environ(), env...)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runner: command failed: %w (output: %s)", err, string(out))
	}
	return nil
}
