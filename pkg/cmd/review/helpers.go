package review

import (
	"fmt"
	"strconv"
)

// parsePRNumber parses and validates a PR number from a string
func parsePRNumber(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid PR number: %s", s)
	}
	if n <= 0 {
		return 0, fmt.Errorf("PR number must be positive: %d", n)
	}
	return n, nil
}
