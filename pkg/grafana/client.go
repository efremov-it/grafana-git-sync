package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"grafana_git_sync/pkg/sync"
)

// Client handles Grafana API operations
type Client struct {
	url      string
	token    string
	user     string
	password string
	client   *http.Client
	folders  map[string]int // cache for folder paths -> IDs
}
func NewClient(url, token, user, password string) *Client {
	return &Client{
		url:      url,
		token:    token,
		user:     user,
		password: password,
		client:   &http.Client{Timeout: 10 * time.Second},
		folders:  make(map[string]int),
	}
}

// WaitForReady waits until Grafana API is available
func (c *Client) WaitForReady(timeout time.Duration) error {
	log.Println("⏳ Waiting for Grafana API...")

	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("%s/api/health", c.url)

	for time.Now().Before(deadline) {
		resp, err := c.client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			log.Println("✅ Grafana API is ready")
			return nil
		}

		if err != nil {
			log.Printf("⚠️ Grafana not ready: %v", err)
		} else {
			log.Printf("⚠️ Grafana returned status: %d", resp.StatusCode)
			resp.Body.Close()
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("Grafana API did not become ready within %v", timeout)
}

// ValidateAuth validates that the provided credentials work
func (c *Client) ValidateAuth() error {
	log.Println("🔐 Validating Grafana credentials...")

	url := fmt.Sprintf("%s/api/health", c.url)

	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(c.user, c.password)

		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("⚠️ Network error: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == 200 {
				log.Println("✅ Grafana authentication OK")
				return nil
			}

			if resp.StatusCode == 401 {
				return fmt.Errorf("Grafana returned 401 Unauthorized — invalid credentials")
			}

			log.Printf("⚠️ Grafana auth returned %d: %s", resp.StatusCode, string(body))
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("Grafana authentication check timed out")
}

// CreateServiceAccountToken creates or recreates a service account token
func (c *Client) CreateServiceAccountToken(accountName, tokenName string) (string, error) {
	// Ensure service account exists
	saID, err := c.ensureServiceAccount(accountName)
	if err != nil {
		return "", fmt.Errorf("cannot ensure service account: %w", err)
	}

	// Create or replace token
	token, err := c.createOrReplaceSAToken(saID, tokenName)
	if err != nil {
		return "", fmt.Errorf("cannot create service account token: %w", err)
	}

	// Validate token works
	if err := c.waitForSAToken(token, 2*time.Minute); err != nil {
		return "", fmt.Errorf("service account token is not ready: %w", err)
	}

	log.Println("✅ Service account token ready")
	return token, nil
}

// GetFolderIDByPath returns the folder ID for a given path from the cache
// Returns 0 if not found
func (c *Client) GetFolderIDByPath(folderPath string) int {
	if folderPath == "" || folderPath == "." {
		return 0
	}
	if id, ok := c.folders[folderPath]; ok {
		return id
	}
	return 0
}

// CreateFolderTree creates a nested folder structure in Grafana
func (c *Client) CreateFolderTree(folderPath string) (int, error) {
	// Check cache first
	if id, ok := c.folders[folderPath]; ok {
		return id, nil
	}

	return c.createFolderRecursive(folderPath, "")
}

// getFolderByTitle searches for an existing folder by title and optional parent UID
func (c *Client) getFolderByTitle(title, parentUid string) (int, string, error) {
	// Use folders API with parentUid parameter to get children of a specific folder
	var url string
	if parentUid == "" {
		// Root level folders
		url = fmt.Sprintf("%s/api/folders?limit=1000", c.url)
	} else {
		// Child folders - use parentUid parameter
		url = fmt.Sprintf("%s/api/folders?parentUid=%s&limit=1000", c.url, parentUid)
	}
	req, _ := http.NewRequest("GET", url, nil)
	c.setAuth(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("failed to list folders: %s", string(body))
	}

	var folders []struct {
		ID        int    `json:"id"`
		UID       string `json:"uid"`
		Title     string `json:"title"`
		ParentUID string `json:"parentUid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return 0, "", err
	}

	// Find folder with matching title and parent
	for _, folder := range folders {
		if folder.Title == title {
			// If we're looking for a root folder (no parent), check parentUid is empty
			if parentUid == "" && folder.ParentUID == "" {
				return folder.ID, folder.UID, nil
			}
			// If we're looking for a subfolder, check parent matches
			if parentUid != "" && folder.ParentUID == parentUid {
				return folder.ID, folder.UID, nil
			}
		}
	}

	return 0, "", nil // Not found
}

// CreateFolderTreeFromNode creates a folder tree from a FolderNode structure
func (c *Client) CreateFolderTreeFromNode(node *sync.FolderNode, parentUid string) error {
	var currentUID string

	// Check if folder already exists in Grafana (always check, even if cached)
	existingID, existingUID, err := c.getFolderByTitle(node.Name, parentUid)
	if err != nil {
		return fmt.Errorf("failed to check existing folder: %w", err)
	}

	if existingID > 0 {
		// Folder already exists, use it
		log.Printf("✅ Folder '%s' already exists (ID: %d, UID: %s, parent: %s)", node.Name, existingID, existingUID, parentUid)
		node.ID = existingID
		node.UID = existingUID
		c.folders[node.FullPath] = existingID
		currentUID = existingUID
	} else {
		// Create new folder
		payload := map[string]string{"title": node.Name}
		if parentUid != "" {
			payload["parentUid"] = parentUid
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/folders", c.url), bytes.NewBuffer(data))
		c.setAuth(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			// Check if error is "folder already exists"
			if resp.StatusCode == 409 || resp.StatusCode == 412 {
				log.Printf("⚠️ Folder '%s' already exists (conflict), fetching it...", node.Name)
				existingID, existingUID, err := c.getFolderByTitle(node.Name, parentUid)
				if err != nil || existingID == 0 {
					return fmt.Errorf("folder exists but cannot retrieve: %s", string(body))
				}
				node.ID = existingID
				node.UID = existingUID
				c.folders[node.FullPath] = existingID
				currentUID = existingUID
			} else {
				return fmt.Errorf("failed to create folder %s: %s", node.Name, string(body))
			}
		} else {
			var created struct {
				ID  int    `json:"id"`
				UID string `json:"uid"`
			}
			if err := json.Unmarshal(body, &created); err != nil {
				return err
			}
			log.Printf("✅ Created folder '%s' (ID: %d, UID: %s, parent: %s)", node.Name, created.ID, created.UID, parentUid)
			node.ID = created.ID
			node.UID = created.UID
			c.folders[node.FullPath] = node.ID
			currentUID = node.UID
		}
	}

	// Process children with the current folder's UID as their parent
	for _, child := range node.Children {
		if err := c.CreateFolderTreeFromNode(child, currentUID); err != nil {
			return err
		}
	}
	return nil
}

// UploadDashboard uploads a dashboard to Grafana
func (c *Client) UploadDashboard(dashboard map[string]interface{}, folderID int) error {
	return c.UploadDashboardWithVersion(dashboard, folderID, "")
}

// UploadDashboardWithVersion uploads a dashboard with version metadata
func (c *Client) UploadDashboardWithVersion(dashboard map[string]interface{}, folderID int, versionMessage string) error {
	body := map[string]interface{}{
		"dashboard": dashboard,
		"folderId":  folderID,
		"overwrite": true,
	}

	// Add version message if provided (links dashboard to Git commit)
	if versionMessage != "" {
		body["message"] = versionMessage
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard JSON: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/dashboards/db", c.url), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Grafana API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) ensureServiceAccount(accountName string) (string, error) {
	req, _ := http.NewRequest("GET", c.url+"/api/serviceaccounts/search", nil)
	req.SetBasicAuth(c.user, c.password)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to list service accounts: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var respData struct {
		ServiceAccounts []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"serviceAccounts"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to parse service accounts list: %w", err)
	}

	for _, sa := range respData.ServiceAccounts {
		if sa.Name == accountName {
			return fmt.Sprintf("%d", sa.ID), nil
		}
	}

	// Create new service account
	payload := fmt.Sprintf(`{"name":"%s","role":"Admin"}`, accountName)
	reqCreate, _ := http.NewRequest("POST", c.url+"/api/serviceaccounts", bytes.NewBuffer([]byte(payload)))
	reqCreate.SetBasicAuth(c.user, c.password)
	reqCreate.Header.Set("Content-Type", "application/json")

	respCreate, err := c.client.Do(reqCreate)
	if err != nil {
		return "", fmt.Errorf("failed to create service account: %w", err)
	}
	defer respCreate.Body.Close()

	bodyCreate, _ := io.ReadAll(respCreate.Body)
	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(bodyCreate, &created); err != nil {
		return "", fmt.Errorf("failed to parse created service account: %w", err)
	}

	log.Println("✅ Service account created:", accountName)
	return fmt.Sprintf("%d", created.ID), nil
}

func (c *Client) createOrReplaceSAToken(saID, tokenName string) (string, error) {
	// List existing tokens
	reqTokens, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/serviceaccounts/%s/tokens", c.url, saID), nil)
	reqTokens.SetBasicAuth(c.user, c.password)
	respTokens, err := c.client.Do(reqTokens)
	if err != nil {
		return "", fmt.Errorf("failed to list tokens: %w", err)
	}
	defer respTokens.Body.Close()

	bodyTokens, _ := io.ReadAll(respTokens.Body)
	var tokensResp []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(bodyTokens, &tokensResp); err != nil {
		return "", fmt.Errorf("failed to parse tokens list: %w", err)
	}

	// Delete old token if exists
	for _, t := range tokensResp {
		if t.Name == tokenName {
			delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/serviceaccounts/%s/tokens/%d", c.url, saID, t.ID), nil)
			delReq.SetBasicAuth(c.user, c.password)
			respDel, err := c.client.Do(delReq)
			if err != nil {
				log.Printf("⚠️ Failed to delete old token %s: %v", tokenName, err)
			} else {
				respDel.Body.Close()
				log.Println("🗑️ Old token deleted:", tokenName)
			}
			break
		}
	}

	// Create new token
	payload := fmt.Sprintf(`{"name":"%s"}`, tokenName)
	reqCreate, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/serviceaccounts/%s/tokens", c.url, saID), bytes.NewBuffer([]byte(payload)))
	reqCreate.SetBasicAuth(c.user, c.password)
	reqCreate.Header.Set("Content-Type", "application/json")

	respCreate, err := c.client.Do(reqCreate)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	defer respCreate.Body.Close()

	bodyCreate, _ := io.ReadAll(respCreate.Body)
	var createdToken struct {
		Key string `json:"key"`
	}

	if err := json.Unmarshal(bodyCreate, &createdToken); err != nil {
		return "", fmt.Errorf("failed to parse created token: %w", err)
	}

	log.Println("✅ Token created:", tokenName)
	return createdToken.Key, nil
}

func (c *Client) waitForSAToken(token string, timeout time.Duration) error {
	url := c.url + "/api/folders"
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("⚠️ Grafana request failed: %v", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == 200 {
				return nil
			} else if resp.StatusCode == 401 || resp.StatusCode == 403 {
				return fmt.Errorf("service account token unauthorized: status %d", resp.StatusCode)
			} else {
				log.Printf("⚠️ Grafana responded with %d, body: %s", resp.StatusCode, string(body))
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("service account token not ready within %v", timeout)
}

func (c *Client) createFolderRecursive(folderPath, parentUID string) (int, error) {
	parts := splitFolderPath(folderPath)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty folder path")
	}

	currentPath := ""
	var folderID int
	var currentUID string = parentUID

	for _, name := range parts {
		if currentPath == "" {
			currentPath = name
		} else {
			currentPath = currentPath + "/" + name
		}

		// Always check if folder exists in Grafana (with correct parent)
		existingID, existingUID, err := c.getFolderByTitle(name, currentUID)
		if err != nil {
			return 0, fmt.Errorf("failed to check existing folder: %w", err)
		}

		if existingID > 0 {
			// Folder already exists, use it
			log.Printf("✅ Folder '%s' already exists (ID: %d, UID: %s, parent: %s)", name, existingID, existingUID, currentUID)
			folderID = existingID
			currentUID = existingUID
			c.folders[currentPath] = folderID
			continue
		}

		// Create new folder
		payload := map[string]string{"title": name}
		if currentUID != "" {
			payload["parentUid"] = currentUID
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/folders", c.url), bytes.NewBuffer(data))
		c.setAuth(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 300 {
			// Check if error is "folder already exists"
			if resp.StatusCode == 409 || resp.StatusCode == 412 {
				log.Printf("⚠️ Folder '%s' already exists (conflict), fetching it...", name)
				existingID, existingUID, err := c.getFolderByTitle(name, currentUID)
				if err != nil || existingID == 0 {
					return 0, fmt.Errorf("folder exists but cannot retrieve: %s", string(body))
				}
				folderID = existingID
				currentUID = existingUID
				c.folders[currentPath] = folderID
				continue
			}
			return 0, fmt.Errorf("failed to create folder %s: %s", name, string(body))
		}

		var created struct {
			ID  int    `json:"id"`
			UID string `json:"uid"`
		}
		if err := json.Unmarshal(body, &created); err != nil {
			return 0, err
		}

		log.Printf("✅ Created folder '%s' (ID: %d, UID: %s, parent: %s)", name, created.ID, created.UID, currentUID)
		folderID = created.ID
		currentUID = created.UID
		c.folders[currentPath] = folderID
	}

	return folderID, nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else {
		req.SetBasicAuth(c.user, c.password)
	}
}

func splitFolderPath(path string) []string {
	if path == "" || path == "." {
		return nil
	}
	parts := []string{}
	for _, p := range splitPath(path) {
		if p != "" && p != "." {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitPath(path string) []string {
	// Support both / and \ as separators
	path = replaceSeparators(path, "/")
	return splitBy(path, "/")
}

func replaceSeparators(s, sep string) string {
	result := ""
	for _, c := range s {
		if c == '/' || c == '\\' {
			result += sep
		} else {
			result += string(c)
		}
	}
	return result
}

func splitBy(s, sep string) []string {
	parts := []string{}
	current := ""
	for _, c := range s {
		if string(c) == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
