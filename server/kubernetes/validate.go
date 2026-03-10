package kubernetes

import (
	"fmt"
	"regexp"
	"strings"
)

var validDeviceIDRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidateDeviceID checks if a device ID is valid for use in K8s resource names.
// The full resource name will be "edge-device-{id}", so total must be <= 253.
// Device IDs must be lowercase — uppercase characters are rejected.
func ValidateDeviceID(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("device ID cannot be empty")
	}
	if len(id) > 240 { // 253 - len("edge-device-") = 241
		return fmt.Errorf("device ID too long: %d chars (max 240)", len(id))
	}
	if id != strings.ToLower(id) {
		return fmt.Errorf("device ID %q must be lowercase", id)
	}
	if !validDeviceIDRegex.MatchString(id) {
		return fmt.Errorf("device ID %q contains invalid characters (must be lowercase alphanumeric or hyphens, cannot start/end with hyphen)", id)
	}
	return nil
}
