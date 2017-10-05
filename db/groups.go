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
func (r *Roach) GroupByUserID(usrID string) ([]model.Group, error) {
	cols := ColDesc(TblGroups+`.`+ColID, TblGroups+`.`+ColName,
		TblGroups+`.`+ColCreateDate, TblGroups+`.`+ColUpdateDate)
	q := `
		SELECT ` + cols + `
			FROM ` + TblUserGroupsJoin + `
			INNER JOIN ` + TblGroups + `
				ON ` + TblUserGroupsJoin + `.` + ColGroupID + `=` + TblGroups + `.` + ColID + `
			WHERE ` + TblUserGroupsJoin + `.` + ColUserID + `=$1`
	rows, err := r.db.Query(q, usrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var grps []model.Group
	for rows.Next() {
		grp := model.Group{}
		err := rows.Scan(&grp.ID, &grp.Name, &grp.CreateDate, &grp.UpdateDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		grps = append(grps, grp)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(grps) == 0 {
		return nil, errors.NewNotFound("no devices found for user")
	}
	return grps, nil
}
