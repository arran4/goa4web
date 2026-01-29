package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGrantExportCmdJSON(t *testing.T) {
	cmd := newGrantExportCmd()

	var out bytes.Buffer
	cmd.fs.SetOutput(&out)

	if err := cmd.Init([]string{"--json"}); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	var payload grantExportPayload
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if len(payload.Sections) == 0 {
		t.Fatal("Expected sections in JSON output")
	}

	section := findGrantExportSection(payload.Sections, "blogs")
	if section == nil {
		t.Fatal("Expected blogs section in JSON output")
	}
	item := findGrantExportItem(section.Items, "entry")
	if item == nil {
		t.Fatal("Expected entry item in JSON output")
	}
	action := findGrantExportAction(item.Actions, "post")
	if action == nil {
		t.Fatal("Expected post action in JSON output")
	}
	if action.Description == "" {
		t.Fatal("Expected description for post action in JSON output")
	}
}

func findGrantExportSection(sections []grantExportSection, name string) *grantExportSection {
	for i := range sections {
		if sections[i].Section == name {
			return &sections[i]
		}
	}
	return nil
}

func findGrantExportItem(items []grantExportItem, name string) *grantExportItem {
	for i := range items {
		if items[i].Item == name {
			return &items[i]
		}
	}
	return nil
}

func findGrantExportAction(actions []grantExportAction, name string) *grantExportAction {
	for i := range actions {
		if actions[i].Action == name {
			return &actions[i]
		}
	}
	return nil
}
