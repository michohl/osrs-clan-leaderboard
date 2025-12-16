package storage

import (
	"database/sql"
	"log"
	"os"
	"slices"
	"strings"

	// https://github.com/mattn/go-sqlite3/issues/335
	_ "github.com/mattn/go-sqlite3"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
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
		position          INTEGER NOT NULL DEFAULT "",
		PRIMARY KEY (server_id, activity)
	);
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("All tables created successfully")
}

// EnrollServer takes form data from our enrollment survey and
// commits that data to our database
func EnrollServer(server model.Servers, activities string) error {
	log.Printf("Request received to enroll server: %s (ID: %s)\n", server.ServerName, server.ID)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	// This design makes it so each server (guild) can only have one
	// channel configured. If we want to allow multiple channels per server
	// then the server ID needs to be a separate column from id
	sqlStmt := table.Servers.
		INSERT(table.Servers.AllColumns).
		MODEL(server).
		ON_CONFLICT(table.Servers.ID).
		DO_UPDATE(
			sqlite.SET(
				table.Servers.ID.SET(sqlite.String(server.ID)),
				table.Servers.ServerName.SET(sqlite.String(server.ServerName)),
				table.Servers.ChannelName.SET(sqlite.String(server.ChannelName)),
				table.Servers.Schedule.SET(sqlite.String(server.Schedule)),
				table.Servers.ShouldEditMessage.SET(sqlite.Bool(server.ShouldEditMessage)),
				table.Servers.IsEnabled.SET(sqlite.Bool(server.IsEnabled)),
			),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	newActivities := strings.Split(activities, ",")

	existingActivityMessages, err := FetchAllMessages(server.ID)
	if err != nil {
		return err
	}

	for _, m := range existingActivityMessages {
		if !slices.Contains(newActivities, m.Activity) {
			RemoveMessage(m)
		}
	}

	for position, activity := range newActivities {

		m, messageErr := FetchMessage(server.ID, activity)

		// If we can't find a message or if the position changed create/update the message row
		if messageErr != nil || int32(position) != m.Position {
			err = EnrollMessage(server, model.Messages{
				MessageID: "",
				ServerID:  server.ID,
				Activity:  activity,
				Position:  int32(position),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnrollUser takes form data from our enrollment survey and
// commits that data to our database
func EnrollUser(user model.Users) error {
	log.Printf("Request received to attach discord user %s to OSRS user %s\n", user.DiscordUsername, user.OsrsUsername)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := table.Users.
		INSERT(table.Users.AllColumns).
		MODEL(user).
		ON_CONFLICT(table.Users.OsrsUsernameKey, table.Users.ServerID).
		DO_UPDATE(
			sqlite.SET(
				table.Users.OsrsUsername.SET(sqlite.String(user.OsrsUsername)),
				table.Users.OsrsAccountType.SET(sqlite.String(user.OsrsAccountType)),
				table.Users.DiscordUsername.SET(sqlite.String(user.DiscordUsername)),
				table.Users.DiscordUserID.SET(sqlite.String(user.DiscordUserID)),
			),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	return nil
}

// RemoveUser removes a user from a specific server
func RemoveUser(user model.Users) error {
	log.Printf("Request received to remove OSRS user %s from server %s\n", user.OsrsUsername, user.ServerID)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := table.Users.
		DELETE().
		WHERE(table.Users.ServerID.
			EQ(sqlite.String(user.ServerID)).
			AND(table.Users.OsrsUsernameKey.EQ(sqlite.String(user.OsrsUsernameKey))),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	return nil
}

// FetchAllServers gets all enrolled servers from the database
func FetchAllServers() ([]model.Servers, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []model.Servers{}, err
	}
	defer db.Close()

	sqlStmt := table.Servers.SELECT(table.Servers.AllColumns)

	var allServers []model.Servers
	err = sqlStmt.Query(db, &allServers)
	if err != nil {
		return allServers, err
	}

	return allServers, nil
}

// FetchServer takes a Guild ID and returns the relevant
// row from our database with the users existing config
func FetchServer(serverID string) (model.Servers, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return model.Servers{}, err
	}
	defer db.Close()

	sqlStmt := table.Servers.
		SELECT(table.Servers.AllColumns).
		WHERE(table.Servers.ID.EQ(sqlite.String(serverID)))

	var s model.Servers
	err = sqlStmt.Query(db, &s)
	if err != nil {
		return model.Servers{}, err
	}

	return s, nil
}

// FetchAllUsers gets all enrolled servers from the database
func FetchAllUsers(serverID string) ([]model.Users, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []model.Users{}, err
	}
	defer db.Close()

	sqlStmt := table.Users.
		SELECT(table.Users.AllColumns).
		WHERE(table.Users.ServerID.EQ(sqlite.String(serverID)))

	var allUsers []model.Users
	sqlStmt.Query(db, &allUsers)

	return allUsers, nil
}

// FetchUser takes a Guild ID and returns the relevant
// row from our database with the users existing config
func FetchUser(serverID string, osrsUsername string) (model.Users, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return model.Users{}, err
	}
	defer db.Close()

	sqlStmt := table.Users.
		SELECT(table.Users.AllColumns).
		WHERE(table.Users.ServerID.
			EQ(sqlite.String(serverID)).
			AND(table.Users.OsrsUsernameKey.EQ(sqlite.String(osrsUsername))),
		)

	u := model.Users{}
	err = sqlStmt.Query(db, &u)
	if err != nil {
		return model.Users{}, err
	}

	return u, nil
}

// FetchAllMessages takes a Guild ID and returns the relevant
// rows from our database that match the Guid ID
func FetchAllMessages(serverID string) ([]model.Messages, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []model.Messages{}, err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		SELECT(table.Messages.AllColumns).
		WHERE(table.Messages.ServerID.EQ(sqlite.String(serverID))).
		ORDER_BY(table.Messages.Position)

	var m []model.Messages
	err = sqlStmt.Query(db, &m)
	if err != nil {
		return []model.Messages{}, err
	}

	return m, nil
}

// FetchMessage takes a Guild ID and returns the relevant
// row from our database that matches the specified activity
func FetchMessage(serverID, activity string) (model.Messages, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return model.Messages{}, err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		SELECT(table.Messages.AllColumns).
		WHERE(table.Messages.ServerID.
			EQ(sqlite.String(serverID)).
			AND(table.Messages.Activity.EQ(sqlite.String(activity))),
		)

	var m model.Messages
	err = sqlStmt.Query(db, &m)
	if err != nil {
		return model.Messages{}, err
	}

	return m, nil
}

// EnrollMessage takes a MessageID for a message we posted
// and stores it so we can update that message later
func EnrollMessage(server model.Servers, message model.Messages) error {
	log.Printf("Request received to enroll message ID %s for server: %s (ID: %s) and activity %s\n", message.MessageID, server.ServerName, server.ID, message.Activity)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		INSERT(table.Messages.AllColumns).
		MODEL(message).
		ON_CONFLICT(table.Messages.ServerID, table.Messages.Activity).
		DO_UPDATE(
			sqlite.SET(
				table.Messages.MessageID.SET(sqlite.String(message.MessageID)),
				table.Messages.ServerID.SET(sqlite.String(server.ID)),
				table.Messages.Activity.SET(sqlite.String(message.Activity)),
				table.Messages.Position.SET(sqlite.Int32(message.Position)),
			),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	return nil
}

// RemoveMessage takes a message and removes the matching row from the database
func RemoveMessage(message model.Messages) error {
	log.Printf("Request received to remove message ID %s for server: %s and activity %s\n", message.MessageID, message.ServerID, message.Activity)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		DELETE().
		WHERE(table.Messages.Activity.
			EQ(sqlite.String(message.Activity)).
			AND(table.Messages.ServerID.EQ(sqlite.String(message.ServerID))).
			AND(table.Messages.MessageID.EQ(sqlite.String(message.MessageID))),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	return nil
}
