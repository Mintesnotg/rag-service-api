package services

import "strings"

// IsConflict detects uniqueness violations based on common driver error messages.
func IsConflict(err error) bool {
    if err == nil {
        return false
    }
    msg := strings.ToLower(err.Error())
    return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
