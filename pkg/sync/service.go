package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Service handles dashboard synchronization
type Service struct {
	repoDir       string
	repoSubdir    string
	dashboardsDir string
	fileHashes    map[string]string // Track file hashes to detect changes
}

// NewService creates a new sync service
func NewService(repoDir, repoSubdir, dashboardsDir string) *Service {
	return &Service{
		repoDir:       repoDir,
		repoSubdir:    repoSubdir,
		dashboardsDir: dashboardsDir,
		fileHashes:    make(map[string]string),
	}
}

// Dashboard represents a dashboard file with its metadata
type Dashboard struct {
	FilePath   string
	FolderPath string
	Content    map[string]interface{}
}

// CopyDashboards copies all JSON dashboard files from the repo to the dashboards directory
func (s *Service) CopyDashboards() ([]string, error) {
	log.Println("📂 Updating dashboards...")
	var updatedFiles []string

	err := filepath.Walk(s.repoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		srcDir := s.repoDir
		if s.repoSubdir != "." && s.repoSubdir != "" {
			srcDir = filepath.Join(s.repoDir, s.repoSubdir)
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(s.dashboardsDir, relPath)

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("❌ Failed to read file %s: %v", path, err)
			return nil
		}

		// Validate JSON
		var tmp interface{}
		if err := json.Unmarshal(content, &tmp); err != nil {
			log.Printf("❌ Invalid JSON in %s: %v", path, err)
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			log.Printf("❌ Failed to create directory for %s: %v", destPath, err)
			return nil
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			log.Printf("❌ Failed to write file %s: %v", destPath, err)
			return nil
		}

		log.Printf("✅ Dashboard updated: %s", destPath)
		updatedFiles = append(updatedFiles, destPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking repo: %w", err)
	}

	return updatedFiles, nil
}

// LoadDashboard reads and parses a dashboard file
func (s *Service) LoadDashboard(filePath string) (*Dashboard, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dashboard file: %w", err)
	}

	var dashboard map[string]interface{}
	if err := json.Unmarshal(content, &dashboard); err != nil {
		return nil, fmt.Errorf("invalid JSON in file: %w", err)
	}

	folderPath := s.detectFolderFromPath(filePath)

	return &Dashboard{
		FilePath:   filePath,
		FolderPath: folderPath,
		Content:    dashboard,
	}, nil
}

// GetUniqueFolders returns a list of unique folder paths from the dashboard files
func (s *Service) GetUniqueFolders(dashboardFiles []string) []string {
	folderSet := make(map[string]bool)
	var folders []string

	for _, file := range dashboardFiles {
		folderPath := s.detectFolderFromPath(file)
		if folderPath != "" && !folderSet[folderPath] {
			folderSet[folderPath] = true
			folders = append(folders, folderPath)
		}
	}

	return folders
}

func (s *Service) detectFolderFromPath(filePath string) string {
	rel, err := filepath.Rel(s.dashboardsDir, filepath.Dir(filePath))
	if err != nil {
		return ""
	}

	rel = filepath.ToSlash(rel)
	if rel == "." || rel == "" {
		return ""
	}

	return rel
}

// FolderNode represents a folder in a tree structure
type FolderNode struct {
	Name     string
	FullPath string
	Children []*FolderNode
	UID      string
	ID       int
}

// BuildFolderGraph creates a tree structure from dashboard file paths
func BuildFolderGraph(dashboardPaths []string, baseDir string) map[string]*FolderNode {
	graph := make(map[string]*FolderNode)

	for _, path := range dashboardPaths {
		rel, _ := filepath.Rel(baseDir, filepath.Dir(path))
		rel = filepath.ToSlash(rel)
		if rel == "." || rel == "" {
			continue
		}

		parts := strings.Split(rel, "/")
		var parentPath string

		for i := 0; i < len(parts); i++ {
			currPath := strings.Join(parts[:i+1], "/")
			if _, ok := graph[currPath]; !ok {
				graph[currPath] = &FolderNode{
					Name:     parts[i],
					FullPath: currPath,
				}
			}
			if i > 0 {
				parentPath = strings.Join(parts[:i], "/")
				// Check if child is not already in Children to avoid duplicates
				found := false
				for _, existingChild := range graph[parentPath].Children {
					if existingChild == graph[currPath] {
						found = true
						break
					}
				}
				if !found {
					graph[parentPath].Children = append(graph[parentPath].Children, graph[currPath])
				}
			}
		}
	}

	return graph
}

// HasParent checks if a node has a parent in the graph
func HasParent(node *FolderNode, graph map[string]*FolderNode) bool {
	for _, n := range graph {
		for _, child := range n.Children {
			if child == node {
				return true
			}
		}
	}
	return false
}

// computeFileHash calculates SHA256 hash of file content
func computeFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// HasFileChanged checks if a file has changed since last sync
func (s *Service) HasFileChanged(path string, content []byte) bool {
	newHash := computeFileHash(content)
	oldHash, exists := s.fileHashes[path]
	
	// If file is new or hash changed, it's been modified
	if !exists || oldHash != newHash {
		s.fileHashes[path] = newHash
		return true
	}
	
	return false
}

// GetChangedFiles returns list of files that changed since last sync
func (s *Service) GetChangedFiles(allFiles []string) ([]string, error) {
	changed := []string{}
	
	for _, filePath := range allFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("⚠️ Failed to read %s: %v", filePath, err)
			continue
		}
		
		if s.HasFileChanged(filePath, content) {
			changed = append(changed, filePath)
		}
	}
	
	return changed, nil
}
