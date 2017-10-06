package db

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

// InsertUserPhone inserts phone details for userID.
func (r *Roach) InsertUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return insertUserPhone(r.db, userID, phone, verified)
}

// InsertUserPhone inserts phone details for userID.
func (r *Roach) InsertUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return insertUserPhone(tx, userID, phone, verified)
}

// UpdateUserPhone updates phone details for userID.
func (r *Roach) UpdateUserPhone(userID, phone string, verified bool) (*model.VerifLogin, error) {
	return updateUserPhone(r.db, userID, phone, verified)
}

// UpdateUserPhoneAtomic updates phone details for userID using tx.
func (r *Roach) UpdateUserPhoneAtomic(tx *sql.Tx, userID, phone string, verified bool) (*model.VerifLogin, error) {
	return updateUserPhone(tx, userID, phone, verified)
}

// InsertPhoneToken persists a token for phone.
func (r *Roach) InsertPhoneToken(userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertPhoneToken(r.db, userID, phone, dbt, isUsed, expiry)
}

// InsertPhoneTokenAtomic persists a token for phone using tx.
func (r *Roach) InsertPhoneTokenAtomic(tx *sql.Tx, userID, phone string, dbt []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	return insertPhoneToken(tx, userID, phone, dbt, isUsed, expiry)
}

// PhoneTokens fetches phone tokens for userID starting with the none-used, newest.
func (r *Roach) PhoneTokens(userID string, offset, count int64) ([]model.DBToken, error) {
	cols := ColDesc(ColID, ColUserID, ColPhone, ColToken, ColIsUsed, ColIssueDate, ColExpiryDate)
	q := `
		SELECT ` + cols + ` FROM ` + TblPhoneTokens + `
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

func insertUserPhone(tx inserter, userID, phone string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: phone, Verified: verified}
	insCols := ColDesc(ColUserID, ColPhone, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblPhones + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, phone, verified).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &vl, nil
}

func updateUserPhone(tx inserter, userID, phone string, verified bool) (*model.VerifLogin, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	vl := model.VerifLogin{UserID: userID, Address: phone, Verified: verified}
	updCols := ColDesc(ColPhone, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	UPDATE ` + TblPhones + `
		SET (` + updCols + `)=($1,$2,CURRENT_TIMESTAMP)
		WHERE ` + ColUserID + `=$3
		RETURNING ` + retCols
	err := tx.QueryRow(q, phone, verified, userID).Scan(&vl.ID, &vl.CreateDate, &vl.UpdateDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("phone for user not found")
		}
		return nil, err
	}
	return &vl, nil
}

func insertPhoneToken(tx inserter, userID, phone string, dbtB []byte, isUsed bool, expiry time.Time) (*model.DBToken, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	dbt := model.DBToken{
		UserID:  userID,
		Address: phone,
		Token:   dbtB,
		IsUsed:  isUsed,
	}
	insCols := ColDesc(ColUserID, ColPhone, ColToken, ColIsUsed, ColExpiryDate)
	retCols := ColDesc(ColID, ColIssueDate, ColExpiryDate)
	q := `
	INSERT INTO ` + TblPhoneTokens + ` (` + insCols + `)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, phone, dbtB, isUsed, expiry).
		Scan(&dbt.ID, &dbt.IssueDate, &dbt.ExpiryDate)
	if err != nil {
		return nil, err
	}
	return &dbt, nil
}
