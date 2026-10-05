package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

const maxSQLUsernameBytes = 32

var compatibleSQLUsername = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// DatabaseUsername adds the TiDB Cloud account prefix and preserves a
// compatible application username when the complete SQL name still fits.
// Otherwise it returns a stable, non-secret identifier within TiDB's limit.
func DatabaseUsername(username, prefix string) string {
	available := maxSQLUsernameBytes - len(prefix)
	if compatibleSQLUsername.MatchString(username) && len(username) <= available {
		return prefix + username
	}
	sum := sha256.Sum256([]byte(strings.ToLower(username)))
	digestLength := available - len("app_")
	return prefix + "app_" + hex.EncodeToString(sum[:])[:digestLength]
}
