package utils

import (
	"os"
	"strings"
)

// IsServerInDevMode takes the name of a discord server and checks if it is
// present in our ENV variable DEV_SERVERS which is used to indicate that we
// are running commands from this server against a local dev copy of this code.
func IsServerInDevMode(server string) bool {
	devWhitelist := strings.SplitSeq(os.Getenv("DEV_SERVERS"), ",")

	for wlServer := range devWhitelist {
		if strings.EqualFold(server, wlServer) {
			return true
		}
	}

	return false
}
