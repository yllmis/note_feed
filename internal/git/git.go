package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func DiffBetween(vaultPath, oldRef, newRef string, filters ...string) (string, error) {
	args := []string{"-C", vaultPath, "diff", oldRef, newRef}
	if len(filters) > 0 {
		args = append(args, "--")
		args = append(args, filters...)
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff 失败: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("执行 git diff 失败: %w", err)
	}
	return string(out), nil
}

func DiffHead(vaultPath string) (string, error) {
	cmd := exec.Command("git", "-C", vaultPath, "diff", "HEAD~1", "HEAD", "--", "*.md")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff HEAD~1 失败: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("执行 git diff HEAD~1 失败: %w", err)
	}
	return string(out), nil
}

func DiffForCommit(vaultPath, commitHash string) (string, error) {
	cmd := exec.Command("git", "-C", vaultPath, "show", commitHash, "--", "*.md")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git show %s 失败: %s", commitHash, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("执行 git show %s 失败: %w", commitHash, err)
	}
	return string(out), nil
}

func GetTodayCommits(vaultPath, date string) ([]string, error) {
	after := date + " 00:00:00"
	before := date + " 23:59:59"
	args := []string{"-C", vaultPath, "log", "--after=" + after, "--before=" + before, "--format=%H", "--no-merges"}
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git log 失败: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("执行 git log 失败: %w", err)
	}
	hashes := strings.Fields(string(out))
	return hashes, nil
}

func IsGitRepo(vaultPath string) bool {
	cmd := exec.Command("git", "-C", vaultPath, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
