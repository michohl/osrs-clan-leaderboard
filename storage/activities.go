package storage

import (
	"database/sql"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
)

// FetchAllActivitiesAndSkills takes a Guild ID and returns all the unique
func FetchAllActivitiesAndSkills(serverID string) ([]string, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return []string{}, err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		SELECT(table.Messages.Activity).
		DISTINCT().
		WHERE(table.Messages.ServerID.EQ(sqlite.String(serverID))).
		ORDER_BY(table.Messages.Position)

	var m []string
	err = sqlStmt.Query(db, &m)
	if err != nil {
		return []string{}, err
	}

	return m, nil

}

// FetchActivityPosition takes a Guild ID and returns all the unique
func FetchActivityPosition(serverID string, activity string) (int64, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	sqlStmt := table.Messages.
		SELECT(table.Messages.Position).
		DISTINCT().
		WHERE(table.Messages.ServerID.
			EQ(sqlite.String(serverID)).
			AND(table.Messages.Activity.EQ(sqlite.String(activity))),
		)

	var m []int64
	err = sqlStmt.Query(db, &m)
	if err != nil {
		return -1, err
	}

	if len(m) > 0 {
		return m[0], nil
	} else {
		return -1, nil
	}
}
