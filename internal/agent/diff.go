package agent

import (
	"os"
	"os/exec"
)

// computeUnifiedDiff produces a unified diff between oldContent and newContent,
// using the system's diff command. filePath is used as the label in the diff header.
// Returns an empty string if the contents are identical or on error.
func computeUnifiedDiff(oldContent, newContent, filePath string) string {
	oldFile, err := os.CreateTemp("", "lif-diff-old-*")
	if err != nil {
		return ""
	}
	defer os.Remove(oldFile.Name())

	newFile, err := os.CreateTemp("", "lif-diff-new-*")
	if err != nil {
		oldFile.Close()
		return ""
	}
	defer os.Remove(newFile.Name())

	_, _ = oldFile.WriteString(oldContent)
	oldFile.Close()

	_, _ = newFile.WriteString(newContent)
	newFile.Close()

	cmd := exec.Command("diff", "-u",
		"--label", "a/"+filePath,
		"--label", "b/"+filePath,
		oldFile.Name(), newFile.Name(),
	)
	output, err := cmd.Output()
	if err != nil {
		// diff exits 1 when files differ — that's expected
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(output)
		}
		// Exit code 2+ means trouble (e.g. binary files), return empty
		return ""
	}
	// Exit code 0 means files are identical
	return ""
}
