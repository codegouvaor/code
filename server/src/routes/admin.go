package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codegouvaor/code/server/src/utils"
)

const dockerSocket = "/var/run/docker.sock"

type dockerContainerInfo struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Image   string            `json:"Image"`
	Labels  map[string]string `json:"Labels"`
}

type dockerLogEntry struct {
	Log    string `json:"log"`
	Stream int    `json:"stream"`
	Time   int64  `json:"time"`
}

func dockerHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", dockerSocket, 5*time.Second)
			},
		},
		Timeout: 10 * time.Second,
	}
}

func dockerGet(path string) ([]byte, error) {
	client := dockerHTTPClient()
	resp, err := client.Get("http://localhost" + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func dockerGetStream(path string) (io.ReadCloser, error) {
	client := dockerHTTPClient()
	resp, err := client.Get("http://localhost" + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("docker API returned %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func mapServiceToContainer(service string) string {
	switch service {
	case "server":
		return "etheriatimes-server"
	case "worker":
		return "etheriatimes-worker"
	case "postgresql", "postgres":
		return "etheriatimes-postgresql"
	case "redis":
		return "etheriatimes-redis"
	case "rabbitmq":
		return "etheriatimes-rabbitmq"
	case "meilisearch":
		return "etheriatimes-meilisearch"
	case "nginx":
		return "etheriatimes-nginx"
	default:
		return "etheriatimes-" + service
	}
}

func findContainerByName(name string) (string, error) {
	data, err := dockerGet("/containers/json?all=true")
	if err != nil {
		return "", err
	}

	var containers []dockerContainerInfo
	if err := json.Unmarshal(data, &containers); err != nil {
		return "", err
	}

	// Try exact match first, then prefix match
	for _, c := range containers {
		for _, n := range c.Names {
			clean := strings.TrimPrefix(n, "/")
			if clean == name || strings.HasPrefix(clean, name) {
				return c.ID[:12], nil
			}
		}
	}

	// Try partial match on service name
	parts := strings.Split(name, "-")
	if len(parts) > 1 {
		serviceKey := parts[len(parts)-1]
		for _, c := range containers {
			for _, n := range c.Names {
				clean := strings.TrimPrefix(n, "/")
				if strings.Contains(clean, serviceKey) {
					return c.ID[:12], nil
				}
			}
		}
	}

	return "", fmt.Errorf("container %q not found", name)
}

// ─── Handlers ───────────────────────────────────────────────────────────────

func (h *apiHandler) listDockerContainers(c *gin.Context) {
	data, err := dockerGet("/containers/json?all=true")
	if err != nil {
		utils.Error(c, utils.NewError(http.StatusServiceUnavailable, "DOCKER_ERROR", "Cannot reach Docker daemon.", nil))
		return
	}

	var raw []dockerContainerInfo
	if err := json.Unmarshal(data, &raw); err != nil {
		utils.Error(c, utils.NewError(http.StatusInternalServerError, "DOCKER_PARSE_ERROR", "Invalid Docker response.", nil))
		return
	}

	containers := make([]gin.H, 0, len(raw))
	for _, r := range raw {
		name := ""
		if len(r.Names) > 0 {
			name = strings.TrimPrefix(r.Names[0], "/")
		}
		containers = append(containers, gin.H{
			"id":     r.ID[:12],
			"name":   name,
			"state":  r.State,
			"status": r.Status,
			"image":  r.Image,
		})
	}

	utils.Success(c, http.StatusOK, containers)
}

func (h *apiHandler) getDockerLogs(c *gin.Context) {
	service := c.Query("service")
	if service == "" {
		service = "server"
	}
	tail := c.DefaultQuery("tail", "100")
	tailNum, err := strconv.Atoi(tail)
	if err != nil || tailNum < 1 {
		tailNum = 100
	}
	if tailNum > 500 {
		tailNum = 500
	}

	containerName := mapServiceToContainer(service)
	containerID, err := findContainerByName(containerName)
	if err != nil {
		utils.Success(c, http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"logs": []string{}},
		})
		return
	}

	path := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true&tail=%d&timestamps=true", containerID, tailNum)
	body, err := dockerGetStream(path)
	if err != nil {
		utils.Success(c, http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"logs": []string{}},
		})
		return
	}
	defer body.Close()

	// Docker multiplexed stream: 8-byte header per frame
	// Bytes 0-3: stream type (1=stdout, 2=stderr)
	// Bytes 4-7: frame size (big-endian uint32)
	raw, _ := io.ReadAll(body)
	logs := parseDockerStream(raw)

	utils.Success(c, http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"logs": logs},
	})
}

func (h *apiHandler) streamDockerLogs(c *gin.Context) {
	service := c.Query("service")
	if service == "" {
		service = "server"
	}

	containerName := mapServiceToContainer(service)
	containerID, err := findContainerByName(containerName)
	if err != nil {
		c.Writer.WriteString("data: {\"error\":\"container not found\"}\n\n")
		c.Writer.Flush()
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	path := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true&tail=20&timestamps=true&follow=true", containerID)
	body, err := dockerGetStream(path)
	if err != nil {
		c.Writer.WriteString("data: {\"error\":\"cannot stream logs\"}\n\n")
		c.Writer.Flush()
		return
	}
	defer body.Close()

	// Read the multiplexed stream frame by frame
	buf := make([]byte, 8)
	for {
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		if _, err := io.ReadFull(body, buf); err != nil {
			return
		}

		size := int(buf[4])<<24 | int(buf[5])<<16 | int(buf[6])<<8 | int(buf[7])
		if size == 0 {
			continue
		}

		payload := make([]byte, size)
		if _, err := io.ReadFull(body, payload); err != nil {
			return
		}

		line := strings.TrimSpace(string(payload))
		if line != "" {
			fmt.Fprintf(c.Writer, "data: %s\n\n", line)
			c.Writer.Flush()
		}
	}
}

func (h *apiHandler) execDockerCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command"`
		Service string `json:"service"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Command) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}

	service := req.Service
	if service == "" {
		service = "server"
	}
	containerName := mapServiceToContainer(service)
	containerID, err := findContainerByName(containerName)
	if err != nil {
		utils.Success(c, http.StatusOK, gin.H{
			"success": false,
			"output":  fmt.Sprintf("Container %q not found", service),
		})
		return
	}

	// Create exec instance
	execBody, _ := json.Marshal(gin.H{
		"AttachStdout": true,
		"AttachStderr": true,
		"Cmd":          []string{"sh", "-c", req.Command},
	})

	createResp, err := dockerPost(fmt.Sprintf("/containers/%s/exec", containerID), execBody)
	if err != nil {
		utils.Success(c, http.StatusOK, gin.H{
			"success": false,
			"output":  fmt.Sprintf("Failed to create exec: %v", err),
		})
		return
	}

	var createResult struct {
		ID string `json:"Id"`
	}
	json.Unmarshal(createResp, &createResult)

	// Start exec
	startBody, _ := json.Marshal(gin.H{"Detach": false, "Tty": false})
	output, err := dockerPost(fmt.Sprintf("/exec/%s/start", createResult.ID), startBody)
	if err != nil {
		utils.Success(c, http.StatusOK, gin.H{
			"success": false,
			"output":  fmt.Sprintf("Failed to execute: %v", err),
		})
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"success": true,
		"output":  string(output),
	})
}

func (h *apiHandler) checkDockerUpdates(c *gin.Context) {
	utils.Success(c, http.StatusOK, gin.H{
		"hasUpdate":      false,
		"currentVersion": "1.0.0",
	})
}

func (h *apiHandler) updateDockerContainer(c *gin.Context) {
	var req struct {
		Image string `json:"image"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Image) == "" {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{
		"success": false,
		"message": "Docker update is not supported in this environment.",
	})
}

func (h *apiHandler) getAdminSettings(c *gin.Context) {
	utils.Success(c, http.StatusOK, gin.H{
		"siteName":          "The Etheria Times",
		"siteDescription":   "L'information au service du citoyen",
		"siteUrl":           "https://etheriatimes.com",
		"email":             "contact@etheriatimes.com",
		"smtpHost":          "",
		"smtpPort":          "587",
		"smtpUser":          "",
		"fromName":          "",
		"fromEmail":         "",
		"maintenanceMode":   false,
		"registrationOpen":  true,
		"commentsEnabled":   true,
		"newsletterEnabled": true,
		"analyticsEnabled":  true,
		"sslEnforced":       true,
	})
}

func (h *apiHandler) updateAdminSettings(c *gin.Context) {
	var req map[string]any
	if c.ShouldBindJSON(&req) != nil {
		utils.Error(c, utils.ErrValidationFailed)
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"updated": true})
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func dockerPost(path string, body []byte) ([]byte, error) {
	client := dockerHTTPClient()
	resp, err := client.Post(
		"http://localhost"+path,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// parseDockerStream parses Docker's multiplexed stdout/stderr stream format.
// Each frame has an 8-byte header: [stream(1), 0, 0, 0, size(4)] followed by
// the payload. We strip the header and return the log lines.
func parseDockerStream(data []byte) []string {
	var lines []string
	i := 0
	for i+8 <= len(data) {
		size := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])
		if size < 0 || i+8+size > len(data) {
			break
		}
		payload := strings.TrimSpace(string(data[i+8 : i+8+size]))
		if payload != "" {
			for _, line := range strings.Split(payload, "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					lines = append(lines, line)
				}
			}
		}
		i += 8 + size
	}
	return lines
}
