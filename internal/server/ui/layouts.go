package ui

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"flowk/internal/config"
)

const maxLayoutPayloadSize = 512 * 1024

type layoutSnapshot struct {
	Version  int                         `json:"version"`
	Viewport *layoutViewport             `json:"viewport,omitempty"`
	Nodes    map[string]layoutCoordinate `json:"nodes"`
}

type layoutViewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

type layoutCoordinate struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type layoutSaveRequest struct {
	FlowID     string         `json:"flowId"`
	SourceName string         `json:"sourceName,omitempty"`
	Snapshot   layoutSnapshot `json:"snapshot"`
}

var layoutTokenPattern = regexp.MustCompile(`[^a-z0-9._-]+`)

func resolveLayoutDir(configPath string) string {
	path := strings.TrimSpace(configPath)
	if path == "" {
		configPath, err := config.ConfigPath()
		if err != nil {
			return ""
		}
		path = configPath
	}
	return filepath.Join(filepath.Dir(path), "ui", "layouts")
}

func layoutFilePath(layoutDir, flowID, sourceName string) (string, error) {
	trimmedID := strings.TrimSpace(flowID)
	if trimmedID == "" {
		return "", errors.New("flowId is required")
	}

	trimmedSource := strings.TrimSpace(sourceName)
	key := trimmedID
	if trimmedSource != "" {
		key = key + "|" + trimmedSource
	}

	base := sanitizeLayoutToken(trimmedID)
	if trimmedSource != "" {
		base = fmt.Sprintf("%s__%s", base, sanitizeLayoutToken(trimmedSource))
	}

	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%s-%x.json", base, hash[:6])
	return filepath.Join(layoutDir, filename), nil
}

func sanitizeLayoutToken(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	sanitized := layoutTokenPattern.ReplaceAllString(normalized, "_")
	sanitized = strings.Trim(sanitized, "._-")
	if sanitized == "" {
		return "flow"
	}
	if len(sanitized) > 64 {
		return sanitized[:64]
	}
	return sanitized
}

func (s *Server) handleGetLayout(c *gin.Context) {
	layoutDir := s.layoutDir
	if strings.TrimSpace(layoutDir) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "layout storage is not configured"})
		return
	}

	flowID := strings.TrimSpace(c.Query("flowId"))
	sourceName := strings.TrimSpace(c.Query("sourceName"))
	path, err := layoutFilePath(layoutDir, flowID, sourceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "layout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read layout"})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

func (s *Server) handleSaveLayout(c *gin.Context) {
	layoutDir := s.layoutDir
	if strings.TrimSpace(layoutDir) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "layout storage is not configured"})
		return
	}

	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLayoutPayloadSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read layout"})
		return
	}

	if len(strings.TrimSpace(string(payload))) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "layout payload is empty"})
		return
	}

	var req layoutSaveRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid layout payload"})
		return
	}

	if strings.TrimSpace(req.FlowID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flowId is required"})
		return
	}

	if req.Snapshot.Version <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid layout version"})
		return
	}

	if req.Snapshot.Nodes == nil {
		req.Snapshot.Nodes = map[string]layoutCoordinate{}
	}

	if req.Snapshot.Viewport != nil {
		if !isFinite(req.Snapshot.Viewport.X) || !isFinite(req.Snapshot.Viewport.Y) || !isFinite(req.Snapshot.Viewport.Zoom) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid viewport"})
			return
		}
	}

	for id, pos := range req.Snapshot.Nodes {
		if strings.TrimSpace(id) == "" {
			delete(req.Snapshot.Nodes, id)
			continue
		}
		if !isFinite(pos.X) || !isFinite(pos.Y) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinates"})
			return
		}
	}

	path, err := layoutFilePath(layoutDir, req.FlowID, req.SourceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := os.MkdirAll(layoutDir, 0o700); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create layouts directory"})
		return
	}

	data, err := json.Marshal(req.Snapshot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not serialize layout"})
		return
	}

	temp, err := os.CreateTemp(layoutDir, "layout-*.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save layout"})
		return
	}

	tempName := temp.Name()
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		_ = os.Remove(tempName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save layout"})
		return
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save layout"})
		return
	}

	if err := os.Rename(tempName, path); err != nil {
		_ = os.Remove(tempName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save layout"})
		return
	}

	if err := os.Chmod(path, 0o600); err != nil {
		_ = os.Remove(path)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save layout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleDeleteLayout(c *gin.Context) {
	layoutDir := s.layoutDir
	if strings.TrimSpace(layoutDir) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "layout storage is not configured"})
		return
	}

	flowID := strings.TrimSpace(c.Query("flowId"))
	sourceName := strings.TrimSpace(c.Query("sourceName"))
	path, err := layoutFilePath(layoutDir, flowID, sourceName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete layout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
