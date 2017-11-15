package db

import (
	"database/sql"

	"github.com/tomogoma/authms/model"
	errors "github.com/tomogoma/go-typed-errors"
)

// InsertGroup inserts into the database returning calculated values.
func (r *Roach) InsertGroup(name string, acl float32) (*model.Group, error) {
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

// Group fetches a group by id.
func (r *Roach) Group(id string) (*model.Group, error) {
	return r.groupWhere(ColID+`=$1`, id)
}

// Group fetches a group by name.
func (r *Roach) GroupByName(name string) (*model.Group, error) {
	return r.groupWhere(ColName+`=$1`, name)
}

// GroupsByUserID fetches a group having id.
func (r *Roach) GroupsByUserID(usrID string) ([]model.Group, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	cols := colDescTbl(TblGroups, ColID, ColName, ColAccessLevel, ColCreateDate, ColUpdateDate)
	q := `
		SELECT ` + cols + `
			FROM ` + TblUserGroupsJoin + `
			INNER JOIN ` + TblGroups + `
				ON ` + TblUserGroupsJoin + `.` + ColGroupID + `=` + TblGroups + `.` + ColID + `
			WHERE ` + TblUserGroupsJoin + `.` + ColUserID + `=$1
			ORDER BY ` + ColAccessLevel + ` ASC
	`
	rows, err := r.db.Query(q, usrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var grps []model.Group
	for rows.Next() {
		grp := model.Group{}
		err := rows.Scan(&grp.ID, &grp.Name, &grp.AccessLevel, &grp.CreateDate, &grp.UpdateDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		grps = append(grps, grp)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(grps) == 0 {
		return nil, errors.NewNotFound("no groups found for user")
	}
	return grps, nil
}

func (r *Roach) Groups(offset, count int64) ([]model.Group, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	cols := ColDesc(ColID, ColName, ColAccessLevel, ColCreateDate, ColUpdateDate)
	q := `
		SELECT ` + cols + ` FROM ` + TblGroups + `
			ORDER BY ` + ColAccessLevel + ` ASC
			LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(q, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var grps []model.Group
	for rows.Next() {
		grp := model.Group{}
		err := rows.Scan(&grp.ID, &grp.Name, &grp.AccessLevel, &grp.CreateDate, &grp.UpdateDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		grps = append(grps, grp)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(grps) == 0 {
		return nil, errors.NewNotFound("no groups found")
	}
	return grps, nil
}

func (r *Roach) groupWhere(where string, whereArgs ...interface{}) (*model.Group, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	cols := ColDesc(ColID, ColName, ColAccessLevel, ColCreateDate, ColUpdateDate)
	q := `SELECT ` + cols + ` FROM ` + TblGroups + ` WHERE ` + where
	grp := model.Group{}
	err := r.db.QueryRow(q, whereArgs...).
		Scan(&grp.ID, &grp.Name, &grp.AccessLevel, &grp.CreateDate, &grp.UpdateDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("groups not found")
		}
		return nil, err
	}
	return &grp, nil
}
