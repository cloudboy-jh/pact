package ui

import (
	"strings"
	"testing"

	"github.com/cloudboy-jh/pact/internal/config"
)

func TestRenderStatusHelpLine(t *testing.T) {
	cfg := &config.PactConfig{Raw: map[string]any{"name": "pact"}}
	output := RenderStatus(cfg, 0, 24)

	if !strings.Contains(output, "[s] sync") {
		t.Fatalf("expected help line to include sync hint")
	}
	if !strings.Contains(output, "[e] edit") {
		t.Fatalf("expected help line to include edit hint")
	}
	if !strings.Contains(output, "[r] refresh") {
		t.Fatalf("expected help line to include refresh hint")
	}
	if !strings.Contains(output, "[j/k] scroll") {
		t.Fatalf("expected help line to include scroll hint")
	}
	if !strings.Contains(output, "[q] quit") {
		t.Fatalf("expected help line to include quit hint")
	}
}
