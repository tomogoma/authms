package db

import (
	"database/sql"

	"github.com/tomogoma/authms/model"
)

func (r *Roach) InsertUserFbIDAtomic(tx *sql.Tx, userID, fbID string, verified bool) (*model.Facebook, error) {
	if tx == nil {
		return nil, errorNilTx
	}
	fb := model.Facebook{UserID: userID, FacebookID: fbID, Verified: verified}
	insCols := ColDesc(ColUserID, ColFacebookID, ColVerified, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblFacebookIDs + ` (` + insCols + `)
		VALUES ($1,$2,$3,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, fbID, verified).Scan(&fb.ID, &fb.CreateDate, &fb.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &fb, nil
}
