package notifier

import (
	"strings"
	"testing"

	"github.com/user/portwatch/internal/config"
)

func TestBuildPayload_ContainsFields(t *testing.T) {
	entry := config.PortEntry{
		Port: 3000,
		Host: "example.com",
	}
	state := "open"
	payload := buildPayload(entry, state)

	checks := []string{"3000", "example.com", "open"}
	for _, c := range checks {
		if !strings.Contains(payload, c) {
			t.Errorf("expected payload to contain %q, got: %s", c, payload)
		}
	}
}

func TestBuildPayload_ValidJSON(t *testing.T) {
	entry := config.PortEntry{
		Port: 443,
		Host: "secure.host",
	}
	payload := buildPayload(entry, "closed")
	if !strings.HasPrefix(payload, "{") || !strings.HasSuffix(payload, "}") {
		t.Errorf("payload does not look like JSON: %s", payload)
	}
}
