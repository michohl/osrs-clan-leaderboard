package types

import "strings"

// OSRSUser is a representation of OSRS Account
type OSRSUser struct {
	Username string
}

// EncodeUsername takes the "human friendly" version
// of the OSRS Username and prepares it for use in
// our API call
func (u OSRSUser) EncodeUsername() string {
	return strings.ReplaceAll(u.Username, " ", "_")
}
