package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/devlyspace/dashspace-cli/internal/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ModuleManifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Providers   []string `json:"providers"`
	Interfaces  []string `json:"interfaces"`
}

type ModuleSearchResult struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
}

func NewClient() *Client {
	cfg := config.GetConfig()
	return &Client{
		baseURL:    cfg.APIBaseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.post("/auth/login", payload)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

func (c *Client) GetCurrentUser() (*User, error) {
	resp, err := c.get("/auth/me")
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateOrGetModule(manifest *ModuleManifest) (int, error) {
	payload := map[string]interface{}{
		"name":         manifest.Name,
		"display_name": manifest.Name,
		"description":  manifest.Description,
		"user_name":    config.GetConfig().Username,
		"visibility":   "public",
	}

	resp, err := c.post("/modules", payload)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}

	if id, ok := result["id"].(float64); ok {
		return int(id), nil
	}

	return 0, fmt.Errorf("unable to get module ID")
}

func (c *Client) SearchModules(query string) ([]ModuleSearchResult, error) {
	endpoint := fmt.Sprintf("/modules?search=%s", query)
	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var searchResponse struct {
		Items []ModuleSearchResult `json:"items"`
		Total int                  `json:"total"`
	}

	if err := json.Unmarshal(resp, &searchResponse); err != nil {
		return nil, err
	}

	return searchResponse.Items, nil
}

func (c *Client) UploadModuleVersion(moduleID int, zipPath string) (int, error) {
	url := fmt.Sprintf("%s/modules/%d/module_versions/upload", c.baseURL, moduleID)

	file, err := os.Open(zipPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return 0, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return 0, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if token := config.GetConfig().AuthToken; token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if id, ok := result["id"].(float64); ok {
		return int(id), nil
	}

	return 0, fmt.Errorf("unable to get version ID")
}

func (c *Client) get(endpoint string) ([]byte, error) {
	return c.request("GET", endpoint, nil)
}

func (c *Client) post(endpoint string, payload interface{}) ([]byte, error) {
	return c.request("POST", endpoint, payload)
}

func (c *Client) request(method, endpoint string, payload interface{}) ([]byte, error) {
	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if token := config.GetConfig().AuthToken; token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
