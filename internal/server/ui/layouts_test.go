package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayoutPersistenceUsesConfigDir(t *testing.T) {
	configHome := t.TempDir()
	configPath := filepath.Join(configHome, "flowk", "config.yaml")

	srv, err := NewServer(Config{
		Address:       "127.0.0.1:0",
		ConfigPath:    configPath,
		FlowUploadDir: filepath.Join(configHome, "uploads"),
	})
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	server := httptest.NewServer(srv.Handle())
	defer server.Close()

	payload := `{"flowId":"root_flow","sourceName":"demo.json","snapshot":{"version":1,"viewport":{"x":10,"y":20,"zoom":1},"nodes":{"task-1":{"x":120,"y":240}}}}`
	resp, err := http.Post(server.URL+"/api/ui/layout", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST layout error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST layout status = %d", resp.StatusCode)
	}

	layoutDir := resolveLayoutDir(configPath)
	layoutPath, err := layoutFilePath(layoutDir, "root_flow", "demo.json")
	if err != nil {
		t.Fatalf("layoutFilePath error: %v", err)
	}

	if _, err := os.Stat(layoutPath); err != nil {
		t.Fatalf("expected layout file at %s: %v", layoutPath, err)
	}

	getResp, err := http.Get(server.URL + "/api/ui/layout?flowId=root_flow&sourceName=demo.json")
	if err != nil {
		t.Fatalf("GET layout error: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET layout status = %d", getResp.StatusCode)
	}

	var snapshot layoutSnapshot
	if err := json.NewDecoder(getResp.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode layout error: %v", err)
	}
	if snapshot.Version != 1 {
		t.Fatalf("layout version = %d", snapshot.Version)
	}
	if snapshot.Nodes["task-1"].X != 120 || snapshot.Nodes["task-1"].Y != 240 {
		t.Fatalf("layout coordinates mismatch: %+v", snapshot.Nodes["task-1"])
	}
}

func TestLayoutDeleteRemovesFile(t *testing.T) {
	configHome := t.TempDir()
	configPath := filepath.Join(configHome, "flowk", "config.yaml")

	srv, err := NewServer(Config{
		Address:       "127.0.0.1:0",
		ConfigPath:    configPath,
		FlowUploadDir: filepath.Join(configHome, "uploads"),
	})
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	server := httptest.NewServer(srv.Handle())
	defer server.Close()

	payload := `{"flowId":"root_flow","sourceName":"demo.json","snapshot":{"version":1,"nodes":{"task-1":{"x":120,"y":240}}}}`
	resp, err := http.Post(server.URL+"/api/ui/layout", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST layout error: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST layout status = %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/ui/layout?flowId=root_flow&sourceName=demo.json", nil)
	if err != nil {
		t.Fatalf("build DELETE request: %v", err)
	}
	deleteResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE layout error: %v", err)
	}
	deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE layout status = %d", deleteResp.StatusCode)
	}

	layoutDir := resolveLayoutDir(configPath)
	layoutPath, err := layoutFilePath(layoutDir, "root_flow", "demo.json")
	if err != nil {
		t.Fatalf("layoutFilePath error: %v", err)
	}
	if _, err := os.Stat(layoutPath); !os.IsNotExist(err) {
		t.Fatalf("expected layout file to be deleted, stat err=%v", err)
	}
}
