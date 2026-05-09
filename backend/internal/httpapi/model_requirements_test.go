package httpapi

import (
	"testing"

	"aigc-backend/internal/modelcfg"
)

func TestModelRequiresInitImage(t *testing.T) {
	h := &Handler{models: sampleVideoModelConfig()}
	if !h.modelRequiresInitImage("sample", "req-image") {
		t.Fatalf("req-image should require init image")
	}
	if h.modelRequiresInitImage("sample", "opt-image") {
		t.Fatalf("opt-image should not require init image")
	}
	if h.modelRequiresInitImage("sample", "text-only") {
		t.Fatalf("text-only should not require init image")
	}
}

func TestModelSupportsInitImage(t *testing.T) {
	h := &Handler{models: sampleVideoModelConfig()}
	if !h.modelSupportsInitImage("sample", "req-image") {
		t.Fatalf("req-image should support init image")
	}
	if !h.modelSupportsInitImage("sample", "opt-image") {
		t.Fatalf("opt-image should support init image")
	}
	if h.modelSupportsInitImage("sample", "text-only") {
		t.Fatalf("text-only should not support init image")
	}
	if !h.modelSupportsInitImage("sample", "wan2.2-i2v-plus") {
		t.Fatalf("i2v heuristic model should support init image")
	}
}

func sampleVideoModelConfig() *modelcfg.Config {
	return &modelcfg.Config{
		Providers: []modelcfg.Provider{
			{
				ID: "sample",
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
