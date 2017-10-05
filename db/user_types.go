package db

import (
	"database/sql"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

// InsertUserType inserts into the database returning calculated values.
func (r *Roach) InsertUserType(name string) (*model.UserType, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	ut := model.UserType{Name: name}
	insCols := ColDesc(ColName, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblUserTypes + ` (` + insCols + `)
		VALUES ($1,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := r.db.QueryRow(q, name).Scan(&ut.ID, &ut.CreateDate, &ut.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &ut, nil
}

func (r *Roach) UserTypeByName(name string) (*model.UserType, error) {
	ut := model.UserType{Name: name}
	cols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `SELECT ` + cols + ` FROM ` + TblUserTypes + ` WHERE ` + ColName + `=$1`
	err := r.db.QueryRow(q, name).Scan(&ut.ID, &ut.CreateDate, &ut.UpdateDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("user type not found")
		}
		return nil, err
	}
	return &ut, nil
}
