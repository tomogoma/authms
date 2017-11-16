package db

import (
	"database/sql"
	"reflect"

	"github.com/lib/pq"
	"github.com/tomogoma/authms/model"
	errors "github.com/tomogoma/go-typed-errors"
	"fmt"
	"strings"
)

var (
	stdUsrCols = ColDesc(
		colDescTbl(TblUsers, ColID, ColPassword, ColCreateDate, ColUpdateDate),
		colDescTbl(TblUserTypes, ColID, ColName, ColCreateDate, ColUpdateDate),
		colDescTbl(TblUserNames, ColID, ColUserName, ColCreateDate, ColUpdateDate),
		colDescTbl(TblEmails, ColID, ColEmail, ColVerified, ColCreateDate, ColUpdateDate),
		colDescTbl(TblPhones, ColID, ColPhone, ColVerified, ColCreateDate, ColUpdateDate),
		colDescTbl(TblFacebookIDs, ColID, ColFacebookID, ColVerified, ColCreateDate, ColUpdateDate),
		colDescTbl(TblGroups, ColID, ColName, ColAccessLevel, ColCreateDate, ColUpdateDate),
	)
)

func (r *Roach) HasUsers(groupID string) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	q := `
		SELECT COUNT(` + ColID + `)
			FROM ` + TblUsers + `
			WHERE ` + ColGroupID + `=$1`
	var numUsers int
	err := r.db.QueryRow(q, groupID).Scan(&numUsers)
	if err != nil {
		return err
	}
	if numUsers == 0 {
		return errors.NewNotFound("No users found")
	}
	return nil
}

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserAtomic(tx *sql.Tx, t model.UserType, g model.Group, password []byte) (*model.User, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	u := model.User{Type: t, Group: g}
	insCols := ColDesc(ColTypeID, ColGroupID, ColPassword, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUsers + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, t.ID, g.ID, password).Scan(&u.ID, &u.CreateDate, &u.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdatePassword stores the new password for userID' account.
func (r *Roach) UpdatePassword(userID string, password []byte) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
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

func (r *Roach) Users(uq model.UsersQuery, offset, count int64) ([]model.User, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}

	qOp := "OR"
	if uq.MatchAll {
		qOp = "AND"
	}
	where := ""
	var whereArgs []interface{}
	i := 1

	if len(uq.GroupNamesIn) > 0 {
		in := "("
		for _, groupName := range uq.GroupNamesIn {
			in = fmt.Sprintf("%s$%d,", in, i)
			whereArgs = append(whereArgs, groupName)
			i++
		}
		in = strings.TrimSuffix(in, ",") + ")"
		where = fmt.Sprintf("%s %s.%s IN %s %s",
			where, TblGroups, ColName, in, qOp)
	}

	if len(uq.ProcessedACLs) > 0 {

		aclOp := "OR"
		if uq.MatchAllACLs {
			aclOp = "AND"
		}

		aclWhere := "("
		for _, aclQ := range uq.ProcessedACLs {
			comp := ""
			// LT and GT are mutually exclusive
			if aclQ.IsLT {
				comp = "<"
			} else if aclQ.IsGT {
				comp = ">"
			}
			// EQ can be exclusive or appended to either LT or GT
			if aclQ.IsEq {
				comp = comp + "="
			}
			// default if no comparator provided
			if comp == "" {
				comp = "="
			}
			aclWhere = fmt.Sprintf("%s %s.%s %s $%d %s",
				aclWhere, TblGroups, ColAccessLevel, comp, i, aclOp)
			whereArgs = append(whereArgs, aclQ.CheckVal)
			i++
		}

		aclWhere = strings.TrimSuffix(aclWhere, aclOp) + ")"
		where = fmt.Sprintf("%s %s %s", where, aclWhere, qOp)
	}

	where = strings.TrimSuffix(where, qOp)
	if where != "" {
		where = fmt.Sprintf("WHERE %s", where)
	}

	limitStr := fmt.Sprintf("$%d", i)
	whereArgs = append(whereArgs, count)
	i++

	offsetStr := fmt.Sprintf("$%d", i)
	whereArgs = append(whereArgs, offset)
	i++

	q := `
		SELECT ` + stdUsrCols + `
			FROM ` + TblUsers + `
			INNER JOIN ` + TblUserTypes + `
				ON ` + TblUsers + `.` + ColTypeID + `=` + TblUserTypes + `.` + ColID + `
			INNER JOIN ` + TblGroups + `
				ON ` + TblUsers + `.` + ColGroupID + `=` + TblGroups + `.` + ColID + `
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
			` + where + `
			ORDER BY ` + TblGroups + `.` + ColAccessLevel + ` ASC
			LIMIT ` + limitStr + ` OFFSET ` + offsetStr + `
	`
	rows, err := r.db.Query(q, whereArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var usrs []model.User
	for rows.Next() {
		usr, _, err := scanStdUser(rows)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		usrs = append(usrs, *usr)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(usrs) == 0 {
		return nil, errors.NewNotFound("no users found")
	}
	return usrs, nil
}

// SetUserGroup associates groupID (from TblGroups) with userID if not
// already associated, otherwise returns an error.
func (r *Roach) SetUserGroup(userID, groupID string) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	q := `UPDATE ` + TblUsers + ` SET ` + ColGroupID + ` = $1 WHERE ` + ColID + ` = $2`
	rslt, err := r.db.Exec(q, userID, groupID)
	if err == sql.ErrNoRows {
		return errors.NewNotFound("user not found")
	}
	return checkRowsAffected(rslt, err, 1)
}

func (r *Roach) userWhere(where string, whereArgs ...interface{}) (*model.User, []byte, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, nil, err
	}
	q := `
	SELECT ` + stdUsrCols + `
		FROM ` + TblUsers + `
			INNER JOIN ` + TblUserTypes + `
				ON ` + TblUsers + `.` + ColTypeID + `=` + TblUserTypes + `.` + ColID + `
			INNER JOIN ` + TblGroups + `
				ON ` + TblUsers + `.` + ColGroupID + `=` + TblGroups + `.` + ColID + `
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

	usr, pass, err := scanStdUser(r.db.QueryRow(q, whereArgs...))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.NewNotFound("user not found")
		}
		return nil, nil, err
	}

	usr.Devices, err = r.UserDevicesByUserID(usr.ID)
	if err != nil && !r.IsNotFoundError(err) {
		return nil, nil, errors.Newf("get device IDs for user: %v", err)
	}

	return usr, pass, nil
}

func scanStdUser(sc scanner) (*model.User, []byte, error) {

	usr := &model.User{}
	var pass []byte
	var usernameID, emailID, phoneID, fbID sql.NullString
	var usernameVal, emailVal, phoneVal, fbVal sql.NullString
	var emailVerified, phoneVerified, fbVerified sql.NullBool
	var usernameCD, emailCD, phoneCD, fbCD pq.NullTime
	var usernameUD, emailUD, phoneUD, fbUD pq.NullTime

	err := sc.Scan(
		&usr.ID, &pass, &usr.CreateDate, &usr.UpdateDate,
		&usr.Type.ID, &usr.Type.Name, &usr.Type.CreateDate, &usr.Type.UpdateDate,
		&usernameID, &usernameVal, &usernameCD, &usernameUD,
		&emailID, &emailVal, &emailVerified, &emailCD, &emailUD,
		&phoneID, &phoneVal, &phoneVerified, &phoneCD, &phoneUD,
		&fbID, &fbVal, &fbVerified, &fbCD, &fbUD, &usr.Group.ID, &usr.Group.Name,
		&usr.Group.AccessLevel, &usr.Group.CreateDate, &usr.Group.UpdateDate,
	)
	if err != nil {
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

	return usr, pass, nil
}

func updatePassword(i inserter, userID string, password []byte) error {
	if i == nil || reflect.ValueOf(i).IsNil() {
		return errorNilTx
	}
	q := `UPDATE ` + TblUsers + ` SET ` + ColPassword + `=$1 WHERE ` + ColID + `=$2`
	res, err := i.Exec(q, password, userID)
	return checkRowsAffected(res, err, 1)
}
