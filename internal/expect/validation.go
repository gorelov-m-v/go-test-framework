package expect

import (
	"fmt"
)

func ValidateJSONPath(path string) error {
	if path == "" {
		return fmt.Errorf("JSON path cannot be empty")
	}
	return nil
}
