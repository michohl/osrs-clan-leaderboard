package types

import "strings"

// OSRSUser is a representation of OSRS Account
type OSRSUser struct {
	// Username is the literal username as it appears in game
	Username string
	// AccountType is the "kind" of account this is. e.g. Ironman
	AccountType string
}

// EncodeUsername takes the "human friendly" version
// of the OSRS Username and prepares it for use in
// our API call
func (u OSRSUser) EncodeUsername() string {
	return strings.ToLower(strings.ReplaceAll(u.Username, " ", "_"))
}
