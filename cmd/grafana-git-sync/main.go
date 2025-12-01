package main

import (
	"fmt"
	"log"
	"time"

	"grafana_git_sync/pkg/config"
	"grafana_git_sync/pkg/git"
	"grafana_git_sync/pkg/grafana"
	"grafana_git_sync/pkg/health"
	"grafana_git_sync/pkg/sync"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Printf("✅ Loaded configuration: %+v\n", cfg.SafeForLog())

	log.Println("🚀 Starting Grafana Git Sync sidecar...")

	// Initialize health checker
	healthChecker := health.NewChecker()

	// Start health check server in background
	go func() {
		if err := healthChecker.StartServer(":8080"); err != nil {
			log.Fatalf("❌ Failed to start health check server: %v", err)
		}
	}()

	// Initialize Grafana client
	grafanaClient := grafana.NewClient(cfg.GrafanaURL, cfg.GrafanaToken, cfg.GrafanaUser, cfg.GrafanaPass)

	// Ensure Grafana is ready
	if err := grafanaClient.WaitForReady(2 * time.Minute); err != nil {
		log.Fatalf("❌ Grafana API not ready: %v", err)
	}
	healthChecker.SetGrafanaHealth(true)

	// Handle service account token creation if needed
	if cfg.GrafanaToken == "" {
		log.Println("ℹ️ No Grafana token provided — creating a new Service Account token...")

		if err := grafanaClient.ValidateAuth(); err != nil {
			log.Fatalf("❌ Grafana authentication failed: %v", err)
		}

		token, err := grafanaClient.CreateServiceAccountToken("git-sync-sa", "git-sync-token")
		if err != nil {
			log.Fatalf("❌ Failed to create service account token: %v", err)
		}

		cfg.GrafanaToken = token
		grafanaClient = grafana.NewClient(cfg.GrafanaURL, token, cfg.GrafanaUser, cfg.GrafanaPass)
		log.Println("✅ Successfully created new Grafana Service Account token")
	} else {
		log.Println("✅ Using provided Grafana Service Account token")
	}

	// Initialize Git client
	gitClient, err := git.NewClient(cfg.RepoURL, cfg.Branch, cfg.RepoDir, cfg.SSHKey, cfg.HTTPSUser, cfg.HTTPSPassword)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Git client: %v", err)
	}

	// Clone repository
	if err := gitClient.Clone(); err != nil {
		log.Fatalf("❌ Failed to clone repository: %v", err)
	}
	healthChecker.SetGitSyncHealth(true)

	// Initialize sync service
	syncService := sync.NewService(cfg.RepoDir, cfg.RepoSubdir, cfg.DashboardsDir)

	var lastCommit string

	// Main sync loop
	for {
		commit, err := gitClient.FetchLatestCommit()
		if err != nil {
			log.Printf("⚠️ Failed to fetch latest commit: %v", err)
			healthChecker.SetLastError(err.Error())
			healthChecker.SetGitSyncHealth(false)
			time.Sleep(cfg.PollInterval)
			continue
		}
		healthChecker.SetGitSyncHealth(true)

		if commit != lastCommit {
			log.Printf("📦 New commit detected: %s", commit)

			// Get commit information for versioning
			commitInfo, err := gitClient.GetCommitInfo()
			if err != nil {
				log.Printf("⚠️ Failed to get commit info: %v", err)
				commitInfo = nil
			}

			// Build version message for Grafana
			versionMessage := ""
			if commitInfo != nil {
				// Format: "commit abc123: Updated dashboard - John Doe"
				shortHash := commitInfo.Hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				versionMessage = fmt.Sprintf("commit %s: %s - %s", shortHash, commitInfo.Message, commitInfo.Author)
				log.Printf("📝 Version: %s", versionMessage)
			}

			// Copy dashboards from repo to dashboards directory
			allFiles, err := syncService.CopyDashboards()
			if err != nil {
				log.Printf("❌ Failed to copy dashboards: %v", err)
				healthChecker.SetLastError(err.Error())
				time.Sleep(cfg.PollInterval)
				continue
			}

		// Smart sync: only process changed files
		changedFiles, err := syncService.GetChangedFiles(allFiles)
		if err != nil {
			log.Printf("⚠️ Failed to detect changed files: %v, syncing all", err)
			changedFiles = allFiles
		}

		if len(changedFiles) == 0 {
			log.Println("ℹ️ No dashboard changes detected in this commit")
			lastCommit = commit
			time.Sleep(cfg.PollInterval)
			continue
		}

		log.Printf("📊 Detected %d changed dashboard(s) out of %d total", len(changedFiles), len(allFiles))

		// Build folder structure (for all files to ensure folders exist)
		folderGraph := sync.BuildFolderGraph(allFiles, cfg.DashboardsDir)

		// Create folders in Grafana (only root nodes, recursively creates children)
		for _, node := range folderGraph {
			if !sync.HasParent(node, folderGraph) {
				if err := grafanaClient.CreateFolderTreeFromNode(node, ""); err != nil {
					log.Printf("❌ Failed to create folder tree %s: %v", node.FullPath, err)
					healthChecker.SetLastError(err.Error())
				}
			}
		}

		// Upload only changed dashboards
		dashboardCount := 0
		for _, filePath := range changedFiles {
			dashboard, err := syncService.LoadDashboard(filePath)
			if err != nil {
				log.Printf("❌ Failed to load dashboard %s: %v", filePath, err)
				continue
			}

			folderID := 0
			if dashboard.FolderPath != "" {
				// Find the folder ID from the already-created folder graph
				folderID = grafanaClient.GetFolderIDByPath(dashboard.FolderPath)
				if folderID == 0 {
					log.Printf("⚠️ Folder not found in graph for %s, attempting to create", dashboard.FolderPath)
					folderID, err = grafanaClient.CreateFolderTree(dashboard.FolderPath)
					if err != nil {
						log.Printf("❌ Failed to ensure folder %s: %v", dashboard.FolderPath, err)
						continue
					}
				}
			}

			if err := grafanaClient.UploadDashboardWithVersion(dashboard.Content, folderID, versionMessage); err != nil {
				log.Printf("❌ Failed to upload dashboard %s: %v", filePath, err)
				healthChecker.SetLastError(err.Error())
			} else {
				log.Printf("✅ Uploaded dashboard: %s", filePath)
				dashboardCount++
			}
		}

		log.Printf("✅ Sync completed: %d dashboard(s) updated", dashboardCount)
		healthChecker.SetLastSync(time.Now())
		healthChecker.SetLastError("")
		lastCommit = commit
		} else {
			log.Println("🔍 No changes detected")
		}

		time.Sleep(cfg.PollInterval)
	}
}
