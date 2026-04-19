package httpapi

import (
	"testing"

	"aigc-backend/internal/modelcfg"
)

func TestModelRequiresInitImage(t *testing.T) {
	h := &Handler{models: mockVideoModelConfig()}
	if !h.modelRequiresInitImage("mock", "req-image") {
		t.Fatalf("req-image should require init image")
	}
	if h.modelRequiresInitImage("mock", "opt-image") {
		t.Fatalf("opt-image should not require init image")
	}
	if h.modelRequiresInitImage("mock", "text-only") {
		t.Fatalf("text-only should not require init image")
	}
}

func TestModelSupportsInitImage(t *testing.T) {
	h := &Handler{models: mockVideoModelConfig()}
	if !h.modelSupportsInitImage("mock", "req-image") {
		t.Fatalf("req-image should support init image")
	}
	if !h.modelSupportsInitImage("mock", "opt-image") {
		t.Fatalf("opt-image should support init image")
	}
	if h.modelSupportsInitImage("mock", "text-only") {
		t.Fatalf("text-only should not support init image")
	}
	if !h.modelSupportsInitImage("mock", "wan2.2-i2v-plus") {
		t.Fatalf("i2v heuristic model should support init image")
	}
}

func mockVideoModelConfig() *modelcfg.Config {
	return &modelcfg.Config{
		Providers: []modelcfg.Provider{
			{
				ID: "mock",
				Video: &modelcfg.CapabilitySpec{
					Models: []modelcfg.ModelSpec{
						{
							ID: "req-image",
							Form: &modelcfg.FormSpec{
								RequiresImage: true,
							},
						},
						{
							ID: "opt-image",
							Form: &modelcfg.FormSpec{
								Fields: []modelcfg.FieldSpec{
									{Key: "prompt", Required: true, Type: "textarea"},
									{Key: "image", Type: "image"},
								},
							},
						},
						{
							ID: "text-only",
							Form: &modelcfg.FormSpec{
								Fields: []modelcfg.FieldSpec{
									{Key: "prompt", Required: true, Type: "textarea"},
								},
							},
						},
					},
				},
			},
		},
	}
}
