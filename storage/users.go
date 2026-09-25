package storage

import (
	"database/sql"
	"log"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
)

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
