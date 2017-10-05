package db

import (
	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

// InsertGroup inserts into the database returning calculated values.
func (r *Roach) InsertGroup(name string, acl int) (*model.Group, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	grp := model.Group{Name: name, AccessLevel: acl}
	insCols := ColDesc(ColName, ColAccessLevel, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblGroups + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := r.db.QueryRow(q, name, acl).Scan(&grp.ID, &grp.CreateDate, &grp.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &grp, nil
}
func (r *Roach) Group(string) (*model.Group, error) {
	return nil, errors.NewNotImplemented()
}
func (r *Roach) GroupByName(string) (*model.Group, error) {
	return nil, errors.NewNotImplemented()
}
