package scanner

import (
	"encoding/json"
	"testing"
)

func TestDetectTechStack_GoProject(t *testing.T) {
	ts := DetectTechStack("../../testdata/go_project")
	if ts.Language != "Go" {
		t.Errorf("language: got %q, want %q", ts.Language, "Go")
	}
	if ts.ProjectName != "go_project" {
		t.Errorf("project name: got %q, want %q", ts.ProjectName, "go_project")
	}
}

func TestDetectTechStack_NodeProject(t *testing.T) {
	ts := DetectTechStack("../../testdata/node_project")
	if ts.Language != "JavaScript/TypeScript" && ts.Language != "TypeScript" {
		t.Errorf("language: got %q, want JavaScript/TypeScript or TypeScript", ts.Language)
	}
	if ts.ProjectName != "node_project" {
		t.Errorf("project name: got %q, want %q", ts.ProjectName, "node_project")
	}
}

func TestDetectTechStack_RailsProject(t *testing.T) {
	ts := DetectTechStack("../../testdata/rails_project")
	if ts.Language != "Ruby" {
		t.Errorf("language: got %q, want %q", ts.Language, "Ruby")
	}
	if ts.Framework != "Rails" {
		t.Errorf("framework: got %q, want %q", ts.Framework, "Rails")
	}
}

func TestDetectTechStack_NonExistentDir(t *testing.T) {
	ts := DetectTechStack("/tmp/memvra-nonexistent-dir")
	if ts.Language != "" {
		t.Errorf("expected empty language for nonexistent dir, got %q", ts.Language)
	}
}

func TestTechStack_JSON(t *testing.T) {
	ts := TechStack{
		ProjectName: "test",
		Language:    "Go",
		Framework:   "stdlib",
	}

	jsonStr := ts.ToJSON()
	if jsonStr == "" {
		t.Fatal("ToJSON returned empty string")
	}

	parsed, err := TechStackFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("TechStackFromJSON error: %v", err)
	}
	if parsed.ProjectName != "test" || parsed.Language != "Go" || parsed.Framework != "stdlib" {
		t.Errorf("round-trip failed: got %+v", parsed)
	}
}

func TestTechStackFromJSON_Invalid(t *testing.T) {
	_, err := TechStackFromJSON("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestTechStack_ToJSON_Valid(t *testing.T) {
	ts := TechStack{
		ProjectName:      "myapp",
		Language:         "Ruby",
		Framework:        "Rails",
		Database:         "PostgreSQL",
		DetectedPatterns: []string{"multi-tenant", "background-jobs"},
	}
	jsonStr := ts.ToJSON()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}
	if m["language"] != "Ruby" {
		t.Errorf("expected language=Ruby, got %v", m["language"])
	}
}
