package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmDeletion prompts the user to confirm deletion of n branches.
// Returns true when the user types "y" or "yes" (case-insensitive). Any
// other answer prints "not a yes -- aborting." and returns false so the
// user knows the input was treated as a decline rather than silently
// dismissed. If stdin is not a terminal, it returns false with nil error.
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
	if interpretConfirm(line) {
		return true, nil
	}
	if _, err := fmt.Fprintln(os.Stdout, "not a yes -- aborting."); err != nil {
		return false, err
	}
	return false, nil
}

func interpretConfirm(line string) bool {
	ans := strings.TrimSpace(strings.ToLower(line))
	return ans == "y" || ans == "yes"
}
