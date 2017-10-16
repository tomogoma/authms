package db

import (
	"database/sql"
	"reflect"

	"github.com/tomogoma/authms/model"
	errors "github.com/tomogoma/go-typed-errors"
)

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserName(userID, username string) (*model.Username, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	return insertUserName(r.db, userID, username)
}

// InsertUserType inserts through tx returning calculated values.
func (r *Roach) InsertUserNameAtomic(tx *sql.Tx, userID, username string) (*model.Username, error) {
	return insertUserName(tx, userID, username)
}

// UpdateUsername sets the new username for userID.
func (r *Roach) UpdateUsername(userID, username string) (*model.Username, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	un := model.Username{UserID: userID, Value: username}
	updCols := ColDesc(ColUserName, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
		UPDATE ` + TblUserNames + ` SET (` + updCols + `)=($1,CURRENT_TIMESTAMP)
			WHERE ` + ColUserID + `=$2
			RETURNING ` + retCols
	err := r.db.QueryRow(q, username, userID).Scan(&un.ID, &un.CreateDate, &un.UpdateDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("username with userID not found")
		}
		return nil, err
	}
	return &un, nil
}

func insertUserName(tx inserter, userID, username string) (*model.Username, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	un := model.Username{UserID: userID, Value: username}
	insCols := ColDesc(ColUserID, ColUserName, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUserNames + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, username).Scan(&un.ID, &un.CreateDate, &un.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &un, nil
}
