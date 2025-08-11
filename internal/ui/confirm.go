package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmDeletion prompts the user to confirm deletion of n branches.
// Returns true when the user types "y" or "yes" (case-insensitive).
// If stdin is not a terminal, it returns false with nil error.
func ConfirmDeletion(n int) (bool, error) {
	// Detect non-interactive stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if (info.Mode() & os.ModeCharDevice) == 0 {
		return false, nil
	}

	reader := bufio.NewReader(os.Stdin)
	if _, err := fmt.Fprintf(os.Stdout, "Proceed with deleting %d branch(es)? [y/N]: ", n); err != nil {
		return false, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	ans := strings.TrimSpace(strings.ToLower(line))
	return ans == "y" || ans == "yes", nil
}
