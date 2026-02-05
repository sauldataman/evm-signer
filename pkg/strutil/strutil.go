package strutil

import (
	"fmt"
	"strconv"
	"strings"
)

// SplitNum parses a range string like "0-9,11,15-20" into a map of indices
// Ranges are inclusive on both ends (0-9 includes 0,1,2,3,4,5,6,7,8,9)
func SplitNum(rangeStr string) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	if rangeStr == "" {
		return result, nil
	}

	parts := strings.Split(rangeStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Handle range like "0-9"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.ParseInt(strings.TrimSpace(rangeParts[0]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid start number in range %s: %v", part, err)
			}

			end, err := strconv.ParseInt(strings.TrimSpace(rangeParts[1]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid end number in range %s: %v", part, err)
			}

			if start > end {
				return nil, fmt.Errorf("invalid range %s: start > end", part)
			}

			for i := start; i <= end; i++ {
				result[strconv.FormatInt(i, 10)] = struct{}{}
			}
		} else {
			// Handle single number
			_, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			result[part] = struct{}{}
		}
	}

	return result, nil
}
