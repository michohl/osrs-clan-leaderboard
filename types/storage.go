package types

// ServersRow is a struct mapping to the fields
// present on every row in the `servers` table
type ServersRow struct {
	ID                int
	ServerName        string
	ChannelName       string
	Activities        string
	Schedule          string
	MessageID         string
	ShouldEditMessage bool
}

// UsersRow is a struct mapping to the fields
// present on every row in the `users` table
type UsersRow struct {
	OSRSUsernameKey string
	OSRSUsername    string
	OSRSAccountType string
	ServerID        string
	DiscordUsername string
	DiscordUserID   int
}
