package storage

import (
	"database/sql"
	"log"
	"os"

	// https://github.com/mattn/go-sqlite3/issues/335
	_ "github.com/mattn/go-sqlite3"
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
        id                  TEXT    NOT NULL PRIMARY KEY,
        server_name         TEXT    NOT NULL DEFAULT "",
		channel_name        TEXT    NOT NULL DEFAULT "",
		schedule            TEXT    NOT NULL DEFAULT "",
		should_edit_message BOOLEAN NOT NULL DEFAULT true,
		is_enabled          BOOLEAN NOT NULL DEFAULT true
    );
    CREATE TABLE IF NOT EXISTS users (
		osrs_username_key TEXT NOT NULL DEFAULT "",
        server_id         TEXT NOT NULL DEFAULT "",
		osrs_username     TEXT NOT NULL DEFAULT "",
		osrs_account_type TEXT NOT NULL DEFAULT "",
		discord_username  TEXT NOT NULL DEFAULT "",
		discord_user_id   TEXT NOT NULL DEFAULT "",
		PRIMARY KEY (osrs_username_key, server_id)
    );
	CREATE TABLE IF NOT EXISTS messages (
		message_id        TEXT    NOT NULL DEFAULT "",
		server_id         TEXT    NOT NULL DEFAULT "",
		activity          TEXT    NOT NULL DEFAULT "",
		position          INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (message_id, server_id, activity)
	);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("All tables created successfully")
}
