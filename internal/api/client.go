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
	"strconv"

	"github.com/devlyspace/devly-cli/internal/config"
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

func NewClient() *Client {
	cfg := config.GetConfig()
	return &Client{
		baseURL:    cfg.APIBaseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	// Note: Adapté selon l'API d'authentification de DashSpace
	// Pour l'instant, simulation - à adapter selon Authenty
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
	// Utilise l'endpoint POST /modules de Modly
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

	// Récupérer l'ID du module créé
	if id, ok := result["id"].(float64); ok {
		return int(id), nil
	}

	return 0, fmt.Errorf("impossible de récupérer l'ID du module")
}

func (c *Client) UploadModuleVersion(moduleID int, zipPath string) (int, error) {
	// Utilise l'endpoint POST /modules/{module_id}/module_versions/upload de Modly
	url := fmt.Sprintf("%s/modules/%d/module_versions/upload", c.baseURL, moduleID)

	// Ouvrir le fichier ZIP
	file, err := os.Open(zipPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Créer le multipart form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Ajouter le fichier
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return 0, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return 0, err
	}

	writer.Close()

	// Créer la requête
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Ajouter l'authentification si disponible
	if token := config.GetConfig().AuthToken; token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Envoyer la requête
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("erreur API (%d): %s", resp.StatusCode, string(body))
	}

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	// Récupérer l'ID de la version
	if id, ok := result["id"].(float64); ok {
		return int(id), nil
	}

	return 0, fmt.Errorf("impossible de récupérer l'ID de la version")
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

	// Ajouter l'authentification si disponible
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
		return nil, fmt.Errorf("erreur API (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
