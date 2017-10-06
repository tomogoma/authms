package db

import (
	"database/sql"
	"reflect"

	"github.com/lib/pq"
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

// UpdatePassword stores the new password for userID' account.
func (r *Roach) UpdatePassword(userID string, password []byte) error {
	return updatePassword(r.db, userID, password)
}

// UpdatePasswordAtomic stores the new password for userID' account using tx.
func (r *Roach) UpdatePasswordAtomic(tx *sql.Tx, userID string, password []byte) error {
	return updatePassword(tx, userID, password)
}

// User fetches User and password for account with id.
func (r *Roach) User(id string) (*model.User, []byte, error) {
	return r.userWhere(TblUsers+`.`+ColID+`=$1`, id)
}

// UserByDeviceID fetches User and password for account with devID.
func (r *Roach) UserByDeviceID(devID string) (*model.User, []byte, error) {
	return r.userWhere(TblDeviceIDs+`.`+ColDevID+`=$1`, devID)
}

// UserByUsername fetches User and password for account with username.
func (r *Roach) UserByUsername(username string) (*model.User, []byte, error) {
	return r.userWhere(TblUserNames+`.`+ColUserName+`=$1`, username)
}

// UserByPhone fetches User and password for account with phone.
func (r *Roach) UserByPhone(phone string) (*model.User, []byte, error) {
	return r.userWhere(TblPhones+`.`+ColPhone+`=$1`, phone)
}

// UserByEmail fetches User and password for account with email.
func (r *Roach) UserByEmail(email string) (*model.User, []byte, error) {
	return r.userWhere(TblEmails+`.`+ColEmail+`=$1`, email)
}

// UserByFacebook fetches User and password for account with fbID.
func (r *Roach) UserByFacebook(fbID string) (*model.User, error) {
	usr, _, err := r.userWhere(TblFacebookIDs+`.`+ColFacebookID+`=$1`, fbID)
	return usr, err
}

// AddUserToGroupAtomic associates groupID (from TblGroups) with userID if not
// already associated, otherwise returns an error.
func (r *Roach) AddUserToGroupAtomic(tx *sql.Tx, userID, groupID string) error {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return errorNilTx
	}
	cols := ColDesc(ColUserID, ColGroupID, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUserGroupsJoin + `(` + cols + `)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (` + ColUserID + `, ` + ColGroupID + `) DO NOTHING
	`
	_, err := tx.Exec(q, userID, groupID)
	return err
}

func (r *Roach) userWhere(where string, whereArgs ...interface{}) (*model.User, []byte, error) {
	cols := ColDesc(
		colDescTbl(TblUsers, ColID, ColPassword, ColCreateDate, ColUpdateDate),
		colDescTbl(TblUserTypes, ColID, ColName, ColCreateDate, ColUpdateDate),
		colDescTbl(TblUserNames, ColID, ColUserName, ColCreateDate, ColUpdateDate),
		colDescTbl(TblEmails, ColID, ColEmail, ColVerified, ColCreateDate, ColUpdateDate),
		colDescTbl(TblPhones, ColID, ColPhone, ColVerified, ColCreateDate, ColUpdateDate),
		colDescTbl(TblFacebookIDs, ColID, ColFacebookID, ColVerified, ColCreateDate, ColUpdateDate),
	)
	q := `
	SELECT ` + cols + `
		FROM ` + TblUsers + `
			INNER JOIN ` + TblUserTypes + `
				ON ` + TblUsers + `.` + ColTypeID + `=` + TblUserTypes + `.` + ColID + `
			LEFT JOIN ` + TblUserNames + `
				ON ` + TblUsers + `.` + ColID + `=` + TblUserNames + `.` + ColUserID + `
			LEFT JOIN ` + TblEmails + `
				ON ` + TblUsers + `.` + ColID + `=` + TblEmails + `.` + ColUserID + `
			LEFT JOIN ` + TblPhones + `
				ON ` + TblUsers + `.` + ColID + `=` + TblPhones + `.` + ColUserID + `
			LEFT JOIN ` + TblFacebookIDs + `
				ON ` + TblUsers + `.` + ColID + `=` + TblFacebookIDs + `.` + ColUserID + `
			LEFT JOIN ` + TblDeviceIDs + `
				ON ` + TblUsers + `.` + ColID + `=` + TblDeviceIDs + `.` + ColUserID + `
		WHERE ` + where

	usr := model.User{}
	var pass []byte
	var usernameID, emailID, phoneID, fbID sql.NullString
	var usernameVal, emailVal, phoneVal, fbVal sql.NullString
	var emailVerified, phoneVerified, fbVerified sql.NullBool
	var usernameCD, emailCD, phoneCD, fbCD pq.NullTime
	var usernameUD, emailUD, phoneUD, fbUD pq.NullTime
	err := r.db.QueryRow(q, whereArgs...).Scan(
		&usr.ID, &pass, &usr.CreateDate, &usr.UpdateDate,
		&usr.Type.ID, &usr.Type.Name, &usr.Type.CreateDate, &usr.Type.UpdateDate,
		&usernameID, &usernameVal, &usernameCD, &usernameUD,
		&emailID, &emailVal, &emailVerified, &emailCD, &emailUD,
		&phoneID, &phoneVal, &phoneVerified, &phoneCD, &phoneUD,
		&fbID, &fbVal, &fbVerified, &fbCD, &fbUD,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.NewNotFound("User not found")
		}
		return nil, nil, err
	}

	if usernameVal.Valid {
		usr.UserName.ID = usernameID.String
		usr.UserName.UserID = usr.ID
		usr.UserName.Value = usernameVal.String
		usr.UserName.CreateDate = usernameCD.Time
		usr.UserName.UpdateDate = usernameUD.Time
	}
	if emailVal.Valid {
		usr.Email.ID = emailID.String
		usr.Email.UserID = usr.ID
		usr.Email.Address = emailVal.String
		usr.Email.Verified = emailVerified.Bool
		usr.Email.CreateDate = emailCD.Time
		usr.Email.UpdateDate = emailUD.Time
	}
	if phoneVal.Valid {
		usr.Phone.ID = phoneID.String
		usr.Phone.UserID = usr.ID
		usr.Phone.Address = phoneVal.String
		usr.Phone.Verified = phoneVerified.Bool
		usr.Phone.CreateDate = phoneCD.Time
		usr.Phone.UpdateDate = phoneUD.Time
	}
	if fbVal.Valid {
		usr.Facebook.ID = fbID.String
		usr.Facebook.UserID = usr.ID
		usr.Facebook.FacebookID = fbVal.String
		usr.Facebook.Verified = fbVerified.Bool
		usr.Facebook.CreateDate = fbCD.Time
		usr.Facebook.UpdateDate = fbUD.Time
	}

	usr.Devices, err = r.UserDevicesByUserID(usr.ID)
	if err != nil && !r.IsNotFoundError(err) {
		return nil, nil, errors.Newf("get device IDs for user: %v", err)
	}

	usr.Groups, err = r.GroupsByUserID(usr.ID)
	if err != nil && !r.IsNotFoundError(err) {
		return nil, nil, errors.Newf("get device IDs for user: %v", err)
	}

	return &usr, pass, nil
}

func updatePassword(i inserter, userID string, password []byte) error {
	if i == nil || reflect.ValueOf(i).IsNil() {
		return errorNilTx
	}
	q := `UPDATE ` + TblUsers + ` SET ` + ColPassword + `=$1 WHERE ` + ColID + `=$2`
	res, err := i.Exec(q, password, userID)
	return checkRowsAffected(res, err, 1)
}
