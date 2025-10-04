package storage

import (
	"database/sql"
	"log"
	"os"

	// https://github.com/mattn/go-sqlite3/issues/335
	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

var (
	// DBFilePath is the user configured path to the SQLite DB file on disk
	DBFilePath = os.Getenv("DB_FILE_PATH")
)

func init() {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS servers (
        id INTEGER NOT NULL PRIMARY KEY,
        server_name TEXT,
		channel_name TEXT,
		tracked_activities TEXT,
		schedule TEXT,
		message_id TEXT DEFAULT "",
		should_edit_message BOOLEAN
    );
    CREATE TABLE IF NOT EXISTS users (
		osrs_username_key TEXT NOT NULL PRIMARY KEY,
		osrs_username TEXT,
		osrs_account_type TEXT DEFAULT "main",
        server_id TEXT,
		discord_username TEXT,
		discord_user_id TEXT
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tables 'servers' and 'users' created successfully")
}

// EnrollServer takes form data from our enrollment survey and
// commits that data to our database
func EnrollServer(server types.ServersRow) error {
	log.Printf("Request received to enroll server: %s (ID: %d)\n", server.ServerName, server.ID)
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// This design makes it so each server (guild) can only have one
	// channel configured. If we want to allow multiple channels per server
	// then the server ID needs to be a separate column from id
	sqlStmt := `
	INSERT OR REPLACE INTO servers (
		id,
		server_name,
		channel_name,
		tracked_activities,
		schedule,
		message_id,
		should_edit_message
	)
	VALUES (
		?,
		?,
		?,
		?,
		?,
		(SELECT message_id from servers where id = ?),
		?
	);
	`
	_, err = db.Exec(
		sqlStmt,
		server.ID,
		server.ServerName,
		server.ChannelName,
		server.Activities,
		server.Schedule,
		server.ID,
		server.ShouldEditMessage,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateServerMessageID takes a MessageID for a message we posted
// and stores it so we can update that message later
func UpdateServerMessageID(server *types.ServersRow, messageID string) error {
	log.Printf("Request received to update message ID %s for server: %s (ID: %d)\n", messageID, server.ServerName, server.ID)
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// This design makes it so each server (guild) can only have one
	// channel configured. If we want to allow multiple channels per server
	// then the server ID needs to be a separate column from id
	sqlStmt := `
	UPDATE servers set message_id = ? WHERE id = ?;
	`
	_, err = db.Exec(
		sqlStmt,
		messageID,
		server.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// EnrollUser takes form data from our enrollment survey and
// commits that data to our database
func EnrollUser(guildID string, discordUser *discordgo.User, osrsUser types.OSRSUser) error {
	log.Printf("Request received to attach discord user %s to OSRS user %s\n", discordUser.Username, osrsUser.Username)
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := `
	INSERT OR REPLACE INTO users (
		osrs_username_key,
		osrs_username,
		osrs_account_type,
		server_id,
		discord_username,
		discord_user_id
	)
	VALUES (
		?,
		?,
		?,
		?,
		?,
		?
	);
	`
	_, err = db.Exec(
		sqlStmt,
		osrsUser.EncodeUsername(), // Normalize input to disallow duplicates
		osrsUser.Username,         // The way the user wants it to appear
		osrsUser.AccountType,
		guildID,
		discordUser.Username,
		discordUser.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// FetchAllServers gets all enrolled servers from the database
func FetchAllServers() ([]*types.ServersRow, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []*types.ServersRow{}, err
	}
	defer db.Close()

	sqlStmt := `SELECT * FROM servers`

	var allServers []*types.ServersRow

	rows, err := db.Query(sqlStmt)
	if err != nil {
		return []*types.ServersRow{}, err
	}
	defer rows.Close()

	for rows.Next() {
		s := &types.ServersRow{}
		err = rows.Scan(
			&s.ID,
			&s.ServerName,
			&s.ChannelName,
			&s.Activities,
			&s.Schedule,
			&s.MessageID,
			&s.ShouldEditMessage,
		)
		if err != nil {
			return []*types.ServersRow{}, err
		}

		allServers = append(allServers, s)
	}

	return allServers, nil
}

// FetchServer takes a Guild ID and returns the relevant
// row from our database with the users existing config
func FetchServer(serverID string) (*types.ServersRow, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return &types.ServersRow{}, err
	}
	defer db.Close()

	sqlStmt := `SELECT * FROM servers WHERE id = ?`
	row := db.QueryRow(sqlStmt, serverID)
	s := &types.ServersRow{}
	err = row.Scan(
		&s.ID,
		&s.ServerName,
		&s.ChannelName,
		&s.Activities,
		&s.Schedule,
		&s.MessageID,
		&s.ShouldEditMessage,
	)
	if err != nil {
		return &types.ServersRow{}, err
	}

	return s, nil
}

// FetchAllUsers gets all enrolled servers from the database
func FetchAllUsers(serverID string) ([]*types.UsersRow, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []*types.UsersRow{}, err
	}
	defer db.Close()

	sqlStmt := `SELECT * FROM users WHERE server_id = ?`

	var allUsers []*types.UsersRow

	rows, err := db.Query(sqlStmt, serverID)
	if err != nil {
		return []*types.UsersRow{}, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &types.UsersRow{}
		err = rows.Scan(
			&u.OSRSUsernameKey,
			&u.OSRSUsername,
			&u.OSRSAccountType,
			&u.ServerID,
			&u.DiscordUsername,
			&u.DiscordUserID,
		)
		if err != nil {
			return []*types.UsersRow{}, err
		}

		allUsers = append(allUsers, u)
	}

	return allUsers, nil
}

// FetchUser takes a Guild ID and returns the relevant
// row from our database with the users existing config
func FetchUser(serverID string, osrsUsername string) (*types.UsersRow, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return &types.UsersRow{}, err
	}
	defer db.Close()

	sqlStmt := `SELECT * FROM users WHERE server_id = ? and osrs_username_key = ?`
	row := db.QueryRow(
		sqlStmt,
		serverID,
		types.OSRSUser{Username: osrsUsername}.EncodeUsername(),
	)
	u := &types.UsersRow{}
	err = row.Scan(
		&u.OSRSUsernameKey,
		&u.OSRSUsername,
		&u.OSRSAccountType,
		&u.DiscordUsername,
		&u.DiscordUserID,
	)
	if err != nil {
		return &types.UsersRow{}, err
	}

	return u, nil
}
