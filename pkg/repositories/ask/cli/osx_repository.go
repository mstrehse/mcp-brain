package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

type OsxRepository struct{}

func (r *OsxRepository) Ask(question string) (string, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf("text returned of (display dialog \"%s\" default answer \"\")", strings.ReplaceAll(question, "\"", "\\\"")))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
