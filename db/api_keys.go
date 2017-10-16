package db

import (
	"github.com/tomogoma/authms/api"
	errors "github.com/tomogoma/go-typed-errors"
)

// InsertAPIKey inserts an API key for the userID.
func (r *Roach) InsertAPIKey(userID, key string) (*api.Key, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	k := api.Key{UserID: userID, APIKey: key}
	insCols := ColDesc(ColUserID, ColKey, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
		INSERT INTO ` + TblAPIKeys + ` (` + insCols + `)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			RETURNING ` + retCols
	err := r.db.QueryRow(q, userID, key).Scan(&k.ID, &k.CreateDate, &k.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// APIKeysByUserID returns API keys for the provided userID starting with the newest.
func (r *Roach) APIKeysByUserID(userID string, offset, count int64) ([]api.Key, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}
	cols := ColDesc(ColID, ColUserID, ColKey, ColCreateDate, ColUpdateDate)
	q := `
	SELECT ` + cols + `
		FROM ` + TblAPIKeys + `
		WHERE ` + ColUserID + `=$1
		ORDER BY ` + ColCreateDate + ` DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(q, userID, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ks []api.Key
	for rows.Next() {
		k := api.Key{}
		err := rows.Scan(&k.ID, &k.UserID, &k.APIKey, &k.CreateDate, &k.UpdateDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		ks = append(ks, k)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(ks) == 0 {
		return nil, errors.NewNotFound("no API Keys found for user")
	}
	return ks, nil
}
