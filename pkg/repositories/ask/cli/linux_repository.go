package cli

import (
	"os/exec"
	"strings"
)

type LinuxRepository struct{}

func (r *LinuxRepository) Ask(question string) (string, error) {
	cmd := exec.Command("zenity", "--entry", "--text", question)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
