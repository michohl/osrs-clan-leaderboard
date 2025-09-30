package storage

import (
	"database/sql"
	"log"

	// https://github.com/mattn/go-sqlite3/issues/335
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbFilePath = "./test.db"
)

func init() {
	db, err := sql.Open("sqlite3", dbFilePath)
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
		schedule TEXT
    );
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        server_name TEXT,
		discord_username TEXT,
		osrs_username TEXT
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
func EnrollServer(server ServersRow) error {
	log.Printf("Request received to enroll server: %s (ID: %d)\n", server.ServerName, server.ID)
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// This design makes it so each server (guild) can only have one
	// channel configured. If we want to allow multiple channels per server
	// then the server ID needs to be a separate column from id
	sqlStmt := `
	INSERT OR REPLACE INTO servers (id, server_name, channel_name, tracked_activities, schedule)
	VALUES (
		?,
		?,
		?,
		?,
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
	)
	if err != nil {
		return err
	}

	return nil
}

// FetchServer takes a Guild ID and returns the relevant
// row from our database with the users existing config
func FetchServer(serverID string) (*ServersRow, error) {

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return &ServersRow{}, err
	}
	defer db.Close()

	sqlStmt := `SELECT id, server_name, channel_name, tracked_activities, schedule FROM servers WHERE id = ?`
	row := db.QueryRow(sqlStmt, serverID)
	s := &ServersRow{}
	err = row.Scan(&s.ID, &s.ServerName, &s.ChannelName, &s.Activities, &s.Schedule)
	if err != nil {
		return &ServersRow{}, err
	}

	return s, nil
}
