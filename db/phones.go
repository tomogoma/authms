package db

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

func (r *Roach) InsertUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return insertUserPhone(r.db, userID, phone, verified)
}
func (r *Roach) InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return insertUserPhone(tx, userID, phone, verified)
}
func (r *Roach) UpdateUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertPhoneToken(r.db, userID, phone, dbt, isUsed, expiry)
}
func (r *Roach) InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertPhoneToken(tx, userID, phone, dbt, isUsed, expiry)
}
func (r *Roach) PhoneTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	return nil, errors.NewNotImplemented()
}

func insertUserPhone(tx inserter, userID, phone string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: phone, Verified: verified}
	insCols := ColDesc(ColUserID, ColPhone, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblPhoneIDs + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, phone, verified).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &vl, nil
}

func insertPhoneToken(tx inserter, userID, phone string, dbtB []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	dbt := model.DBToken{
		UserID:     userID,
		Address:    phone,
		Token:      dbtB,
		IsUsed:     isUsed,
		ExpiryDate: expiry,
	}
	insCols := ColDesc(ColUserID, ColPhone, ColToken, ColIsUsed, ColExpiryDate)
	retCols := ColDesc(ColID, ColIssueDate)
	q := `
	INSERT INTO ` + TblPhoneTokens + ` (` + insCols + `)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, phone, dbtB, isUsed, expiry).
		Scan(&dbt.ID, &dbt.IssueDate)
	if err != nil {
		return nil, err
	}
	return &dbt, nil
}
