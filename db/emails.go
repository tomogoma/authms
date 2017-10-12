package db

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

// InsertUserPhone inserts email details for userID.
func (r *Roach) InsertUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	return insertUserEmail(r.db, userID, email, verified)
}

// InsertUserEmailAtomic inserts email details for userID.
func (r *Roach) InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return insertUserEmail(tx, userID, email, verified)
}

// UpdateUserEmail updates email details for userID.
func (r *Roach) UpdateUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	return updateUserEmail(r.db, userID, email, verified)
}

// UpdateUserEmailAtomic updates email details for userID using tx.
func (r *Roach) UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return updateUserEmail(tx, userID, email, verified)
}

// InsertEmailToken persists a token for email.
func (r *Roach) InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	return insertEmailToken(r.db, userID, email, dbt, isUsed, expiry)
}

// InsertEmailTokenAtomic persists a token for email using tx.
func (r *Roach) InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertEmailToken(tx, userID, email, dbt, isUsed, expiry)
}

// EmailTokens fetches email tokens for userID starting with the newest.
func (r *Roach) EmailTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	cols := ColDesc(ColID, ColUserID, ColEmail, ColToken, ColIsUsed, ColIssueDate, ColExpiryDate)
	q := `
		SELECT ` + cols + ` FROM ` + TblEmailTokens + `
			WHERE ` + ColUserID + `=$1
			ORDER BY ` + ColIsUsed + ` ASC, ` + ColIssueDate + ` DESC
			LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(q, userID, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbts []model.DBToken
	for rows.Next() {
		dbt := model.DBToken{}
		err := rows.Scan(&dbt.ID, &dbt.UserID, &dbt.Address, &dbt.Token,
			&dbt.IsUsed, &dbt.IssueDate, &dbt.ExpiryDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		dbts = append(dbts, dbt)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(dbts) == 0 {
		return nil, errors.NewNotFound("no devices found for user")
	}
	return dbts, nil
}

func insertUserEmail(tx inserter, userID, address string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: address, Verified: verified}
	insCols := ColDesc(ColUserID, ColEmail, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblEmails + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, address, verified).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &vl, nil
}

func updateUserEmail(tx inserter, userID, address string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: address, Verified: verified}
	updCols := ColDesc(ColEmail, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	UPDATE ` + TblEmails + `
		SET (` + updCols + `)=($1,$2,CURRENT_TIMESTAMP)
		WHERE ` + ColUserID + `=$3
		RETURNING ` + retCols
	err := tx.QueryRow(q, address, verified, userID).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("email for user not found")
		}
		return nil, err
	}
	return &vl, nil
}

func insertEmailToken(tx inserter, userID, address string, dbtB []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	dbt := model.DBToken{
		UserID:  userID,
		Address: address,
		Token:   dbtB,
		IsUsed:  isUsed,
	}
	insCols := ColDesc(ColUserID, ColEmail, ColToken, ColIsUsed, ColExpiryDate)
	retCols := ColDesc(ColID, ColIssueDate, ColExpiryDate)
	q := `
	INSERT INTO ` + TblEmailTokens + ` (` + insCols + `)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, address, dbtB, isUsed, expiry).
		Scan(&dbt.ID, &dbt.IssueDate, &dbt.ExpiryDate)
	if err != nil {
		return nil, err
	}
	return &dbt, nil
}
