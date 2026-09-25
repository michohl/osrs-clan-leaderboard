package storage

import (
	"database/sql"
	"log"
	"slices"
	"strings"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
)

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

		currentPosition, err := FetchActivityPosition(server.ID, activity)

		// If we can't find a message or if the position changed create/update the message row
		if err != nil || currentPosition == -1 || int64(position) != currentPosition {
			err = EnrollMessage(server, model.Messages{
				MessageID: "",
				ServerID:  server.ID,
				Activity:  activity,
				Position:  int64(position),
			})
			if err != nil {
				return err
			}
		}
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
