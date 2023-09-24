package repository_game

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain/model"
	"github.com/stretchr/testify/assert"
)

func initMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGameRepo_Get(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	gameID := 0
	rows := sqlmock.NewRows([]string{
		"id",
		"white_id",
		"black_id",
		"time",
		"increment",
		"result",
		"winner",
		"version",
		"time_stamp_at_turn_start",
		"white_time",
		"black_time",
		"history",
		"moves",
		"black",
		"white",
	}).
		AddRow(gameID, "faa", "fab", 5000, 0, "", "", 0, time.Now().Unix(), 5000, 5000, "", "", false, true)

	query := fmt.Sprintf(`
        SELECT 
            game.*,
            drawrecord.black,
            drawrecord.white
        FROM 
            game g
        LEFT JOIN 
            drawrecord dr
        ON
            g.id = dr.game_id
        WHERE 
            g.id = $1
    `,
	)

	prep := mock.ExpectPrepare(query)
	prep.ExpectQuery().WillReturnRows(rows)

	r := NewGameRepo(db)

	game, err := r.Get(context.Background(), gameID)

	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.NotNil(t, game.DrawRecord)
	assert.True(t, game.DrawRecord.White)
}

func TestGameRepo_Update(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	gameID := 0
	mockGame := new(domain.Game)

	err := faker.FakeData(mockGame)
	assert.NoError(t, err)

	newVersion := mockGame.Version + 1
	newWhiteTime := 50000000

	query := fmt.Sprintf(`
    UPDATE game 
    SET 
        version = %d,
        white_time = %d
    WHERE
        id = %d
    AND
        version = %d
    `,
		newVersion,
		newWhiteTime,
		gameID,
		mockGame.Version,
	)

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameRepo(db)

	changes := make(map[string]interface{})
	changes["WhiteTime"] = newWhiteTime

	updated, err := r.Update(context.Background(), gameID, mockGame.Version, changes)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestGameRepo_Insert(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	query := fmt.Sprintf(`
    INSERT INTO game (
        white_id,
        black_id,
        time,
        increment,
        version,
        time_stamp_at_turn_start,
        white_time,
        black_time
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8
    )`,
	)
	prep := mock.ExpectPrepare(query)

	whiteID := "four"
	blackID := "five"
	timeData := 5000
	increment := 5
	version := 1
	timeStampAtTurnStart := time.Now().Unix()
	whiteTime := timeData
	blackTime := timeData
	prep.ExpectExec().WithArgs(
		whiteID,
		blackID,
		timeData,
		increment,
		version,
		timeStampAtTurnStart,
		whiteTime,
		blackTime,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameRepo(db)

	err := r.Insert(
		context.Background(),
		&domain.Game{
			WhiteID:              whiteID,
			BlackID:              blackID,
			Time:                 timeData,
			Increment:            increment,
			Version:              version,
			TimeStampAtTurnStart: timeStampAtTurnStart,
			WhiteTime:            whiteTime,
			BlackTime:            blackTime,
		})

	assert.NoError(t, err)
}