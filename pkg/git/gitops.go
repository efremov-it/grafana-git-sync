package git

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

// CommitInfo contains metadata about a Git commit
type CommitInfo struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Timestamp time.Time
}

// Client handles Git operations
type Client struct {
	repoURL  string
	branch   string
	repoDir  string
	auth     transport.AuthMethod
	repo     *gogit.Repository
}

// NewClient creates a new Git client with authentication
func NewClient(repoURL, branch, repoDir, sshKey, httpsUser, httpsPassword string) (*Client, error) {
	auth, err := createAuth(repoURL, sshKey, httpsUser, httpsPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create git auth: %w", err)
	}

	return &Client{
		repoURL: repoURL,
		branch:  branch,
		repoDir: repoDir,
		auth:    auth,
	}, nil
}

// Clone clones the repository to the local directory
func (c *Client) Clone() error {
	log.Println("📥 Cloning repo...")

	// Remove old repo directory if exists
	if err := os.RemoveAll(c.repoDir); err != nil {
		return fmt.Errorf("failed to remove old repo directory: %w", err)
	}

	repo, err := gogit.PlainClone(c.repoDir, false, &gogit.CloneOptions{
		URL:           c.repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(c.branch),
		SingleBranch:  true,
		Depth:         1,
		Auth:          c.auth,
		Progress:      os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	c.repo = repo
	log.Println("✅ Repo cloned successfully")
	return nil
}

// FetchLatestCommit pulls the latest changes and returns the commit hash
func (c *Client) FetchLatestCommit() (string, error) {
	if c.repo == nil {
		return "", fmt.Errorf("repository not initialized, call Clone first")
	}

	w, err := c.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Pull(&gogit.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(c.branch),
		SingleBranch:  true,
		Force:         true,
		Auth:          c.auth,
	})

	if err != nil && err != gogit.NoErrAlreadyUpToDate && !strings.Contains(err.Error(), "empty git-upload-pack") {
		return "", fmt.Errorf("pull failed: %w", err)
	}

	ref, err := c.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return ref.Hash().String(), nil
}

// GetCommitInfo returns detailed information about the current HEAD commit
func (c *Client) GetCommitInfo() (*CommitInfo, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	ref, err := c.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := c.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	return &CommitInfo{
		Hash:      commit.Hash.String(),
		Message:   strings.TrimSpace(commit.Message),
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Timestamp: commit.Author.When,
	}, nil
}

func createAuth(repoURL, sshKey, httpsUser, httpsPassword string) (transport.AuthMethod, error) {
	// SSH authentication
	if sshKey != "" && (strings.HasPrefix(repoURL, "ssh://") || strings.HasPrefix(repoURL, "git@")) {
		keyStr := strings.ReplaceAll(sshKey, `\n`, "\n")

		if strings.TrimSpace(keyStr) == "" {
			return nil, fmt.Errorf("SSH key is empty")
		}

		signer, err := ssh.ParsePrivateKey([]byte(keyStr))
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %w", err)
		}

		log.Printf("🔑 SSH key successfully parsed, length: %d bytes", len(keyStr))

		auth := &gitssh.PublicKeys{
			User:   "git",
			Signer: signer,
			HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}
		log.Println("✅ Using SSH auth with key from environment")
		return auth, nil
	}

	// HTTPS authentication
	if httpsUser != "" && httpsPassword != "" {
		log.Println("🔑 Using HTTPS auth")
		return &githttp.BasicAuth{
			Username: httpsUser,
			Password: httpsPassword,
		}, nil
	}

	return nil, fmt.Errorf("no authentication provided")
}
