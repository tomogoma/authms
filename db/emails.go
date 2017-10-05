package db

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

func (r *Roach) InsertUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return insertUserEmail(r.db, userID, email, verified)
}
func (r *Roach) InsertUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return insertUserEmail(tx, userID, email, verified)
}
func (r *Roach) UpdateUserEmail(userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserEmailAtomic(tx *sql.Tx, userID, email string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertEmailToken(userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertEmailToken(r.db, userID, email, dbt, isUsed, expiry)
}
func (r *Roach) InsertEmailTokenAtomic(tx *sql.Tx, userID, email string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertEmailToken(tx, userID, email, dbt, isUsed, expiry)
}
func (r *Roach) EmailTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}

func insertUserEmail(tx inserter, userID, address string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: address, Verified: verified}
	insCols := ColDesc(ColUserID, ColEmail, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblEmailIDs + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, address, verified).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &vl, nil
}

func insertEmailToken(tx inserter, userID, address string, dbtB []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	dbt := model.DBToken{
		UserID:     userID,
		Address:    address,
		Token:      dbtB,
		IsUsed:     isUsed,
		ExpiryDate: expiry,
	}
	insCols := ColDesc(ColUserID, ColEmail, ColToken, ColIsUsed, ColExpiryDate)
	retCols := ColDesc(ColID, ColIssueDate)
	q := `
	INSERT INTO ` + TblEmailTokens + ` (` + insCols + `)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, address, dbtB, isUsed, expiry).
		Scan(&dbt.ID, &dbt.IssueDate)
	if err != nil {
		return nil, err
	}
	return &dbt, nil
}
