package storage

import (
	"database/sql"
	"log"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
)

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
func FetchMessage(serverID, activity string) ([]model.Messages, error) {

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []model.Messages{}, err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		SELECT(table.Messages.AllColumns).
		WHERE(table.Messages.ServerID.
			EQ(sqlite.String(serverID)).
			AND(table.Messages.Activity.EQ(sqlite.String(activity))),
		)

	var m []model.Messages
	err = sqlStmt.Query(db, &m)
	if err != nil {
		return []model.Messages{}, err
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
		ON_CONFLICT(table.Messages.MessageID, table.Messages.ServerID, table.Messages.Activity).
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
