package db

import (
	"database/sql"
	"reflect"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserAtomic(tx *sql.Tx, t model.UserType, password []byte) (*model.User, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	u := model.User{Type: t}
	insCols := ColDesc(ColTypeID, ColPassword, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUsers + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, t.ID, password).Scan(&u.ID, &u.CreateDate, &u.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Roach) UpdatePassword(userID string, password []byte) error {
	return updatePassword(r.db, userID, password)
}

func (r *Roach) UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error {
	return updatePassword(tx, userID, password)
}

func (r *Roach) User(id string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByDeviceID(devID string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByUsername(username string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByPhone(phone string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByEmail(email string) (*model.User, []byte, error) {
	return nil, nil, errors.NewNotImplemented()
}
func (r *Roach) UserByFacebook(facebookID string) (*model.User, error) {
	return nil, errors.NewNotImplemented()
}

func (r *Roach) AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error {
	return errors.NewNotImplemented()
}

func updatePassword(i inserter, userID string, password []byte) error {
	if i == nil || reflect.ValueOf(i).IsNil() {
		return errorNilTx
	}
	q := `UPDATE ` + TblUsers + ` SET ` + ColPassword + `=$1 WHERE ` + ColID + `=$2`
	res, err := i.Exec(q, password, userID)
	return checkRowsAffected(res, err, 1)
}
