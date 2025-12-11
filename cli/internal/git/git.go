package git

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Clone clones the user's my-pact repo to ~/.pact/
func Clone(token, username string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pactDir := filepath.Join(home, ".pact")

	// Remove existing directory if it exists
	if _, err := os.Stat(pactDir); err == nil {
		if err := os.RemoveAll(pactDir); err != nil {
			return fmt.Errorf("failed to remove existing .pact directory: %w", err)
		}
	}

	repoURL := fmt.Sprintf("https://github.com/%s/my-pact.git", username)

	_, err = git.PlainClone(pactDir, false, &git.CloneOptions{
		URL: repoURL,
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	return nil
}

// Pull pulls the latest changes from the remote
func Pull(token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pactDir := filepath.Join(home, ".pact")

	repo, err := git.PlainOpen(pactDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Progress: os.Stdout,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// Push commits and pushes local changes to the remote
func Push(token, message string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pactDir := filepath.Join(home, ".pact")

	repo, err := git.PlainOpen(pactDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check for changes
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return fmt.Errorf("no changes to commit")
	}

	// Stage all changes
	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Get user info from git config
	cfg, err := repo.Config()
	if err != nil {
		cfg = &config.Config{}
	}

	authorName := cfg.User.Name
	authorEmail := cfg.User.Email
	if authorName == "" {
		authorName = "pact"
	}
	if authorEmail == "" {
		authorEmail = "pact@users.noreply.github.com"
	}

	// Commit
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// HasChanges checks if there are uncommitted changes
func HasChanges() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	pactDir := filepath.Join(home, ".pact")

	repo, err := git.PlainOpen(pactDir)
	if err != nil {
		return false, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

// GetStatus returns the git status of the pact repo
func GetStatus() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	pactDir := filepath.Join(home, ".pact")

	repo, err := git.PlainOpen(pactDir)
	if err != nil {
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	status, err := worktree.Status()
	if err != nil {
		return "", err
	}

	return status.String(), nil
}
