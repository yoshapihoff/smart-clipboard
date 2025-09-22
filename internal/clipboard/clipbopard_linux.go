//go:build linux
// +build linux

package clipboard

import (
	"os/exec"
	"strings"
)

func GetClipboard() (string, error) {
	// Попробуем xclip
	cmd := exec.Command("xclip", "-o", "-selection", "clipboard")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Попробуем xsel
	cmd = exec.Command("xsel", "--clipboard", "--output")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "", err
}

func SetClipboard(content string) error {
	// Попробуем xclip
	cmd := exec.Command("xclip", "-i", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Попробуем xsel
	cmd = exec.Command("xsel", "--clipboard", "--input")
	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}
