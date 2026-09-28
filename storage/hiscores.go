package storage

import (
	"database/sql"
	"log"

	"github.com/go-jet/jet/v2/sqlite"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/table"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// EnrollHiscores makes note of all skills/activities for a user and what xp/rank they had at the time for later comparison
func EnrollHiscores(server model.Servers, hiscores model.SeenHiscores) error {
	log.Printf("Request received to capture the %s hiscores of user %s on server: %s (ID: %s)\n", hiscores.Activity, hiscores.OsrsUsernameKey, server.ServerName, server.ID)

	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := table.SeenHiscores.
		INSERT(table.SeenHiscores.AllColumns).
		MODEL(hiscores).
		ON_CONFLICT(table.SeenHiscores.OsrsUsernameKey, table.SeenHiscores.ServerID, table.SeenHiscores.Activity).
		DO_UPDATE(
			sqlite.SET(
				table.SeenHiscores.Rank.SET(sqlite.Int64(hiscores.Rank)),
				table.SeenHiscores.Xp.SET(sqlite.Int64(hiscores.Xp)),
				table.SeenHiscores.Score.SET(sqlite.Int64(hiscores.Score)),
				table.SeenHiscores.Level.SET(sqlite.Int64(hiscores.Level)),
			),
		)

	_, err = sqlStmt.Exec(db)
	if err != nil {
		return err
	}

	return nil
}

// FetchPreviousHiscore takes a Guild ID + username + activity and returns the relevant
// row from our database that matches
func FetchPreviousHiscore(serverID, username, activity string) (model.SeenHiscores, error) {
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return model.SeenHiscores{}, err
	}
	defer db.Close()

	sqlStmt := table.SeenHiscores.
		SELECT(table.SeenHiscores.AllColumns).
		WHERE(table.SeenHiscores.ServerID.
			EQ(sqlite.String(serverID)).
			AND(table.SeenHiscores.OsrsUsernameKey.EQ(sqlite.String(username))).
			AND(table.SeenHiscores.Activity.EQ(sqlite.String(activity))),
		)

	var hs model.SeenHiscores
	err = sqlStmt.Query(db, &hs)
	if err != nil {
		return model.SeenHiscores{}, err
	}

	return hs, nil
}

func EnrollManyHiscores(server model.Servers, allHiscores map[model.Users]types.Hiscores) error {

	for user, hiscores := range allHiscores {
		for _, activity := range hiscores.Activities {
			hs := model.SeenHiscores{
				OsrsUsernameKey: user.OsrsUsernameKey,
				ServerID:        user.ServerID,
				Activity:        activity.Name,
				Rank:            int64(activity.Rank),
				Score:           int64(activity.Score),
				Xp:              0,
			}
			EnrollHiscores(server, hs)
		}

		for _, activity := range hiscores.Skills {
			hs := model.SeenHiscores{
				OsrsUsernameKey: user.OsrsUsernameKey,
				ServerID:        user.ServerID,
				Activity:        activity.Name,
				Rank:            int64(activity.Rank),
				Xp:              int64(activity.XP),
				Level:           int64(activity.Level),
				Score:           0,
			}
			EnrollHiscores(server, hs)
		}
	}

	return nil
}
