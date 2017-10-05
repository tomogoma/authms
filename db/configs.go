package db

import (
	"database/sql"
	"encoding/json"

	"github.com/tomogoma/go-commons/errors"
)

// UpsertSMTPConfig upserts SMTP config values into the db.
func (r *Roach) UpsertSMTPConfig(conf interface{}) error {
	return r.upsertConf(keySMTPConf, conf)
}

// GetSMTPConfig fetches SMTP config values from the db and unmarshals them
// into conf. this method fails if conf is nil or not a pointer.
func (r *Roach) GetSMTPConfig(conf interface{}) error {
	return r.getConf(keySMTPConf, conf)
}

func (r *Roach) upsertConf(key string, conf interface{}) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	valB, err := json.Marshal(conf)
	if err != nil {
		return errors.Newf("marshal conf: %v", err)
	}
	cols := ColDesc(ColKey, ColValue, ColUpdateDate)
	q := `UPSERT INTO ` + TblConfigurations + ` (` + cols + `) VALUES ($1, $2, CURRENT_TIMESTAMP)`
	res, err := r.db.Exec(q, key, valB)
	return checkRowsAffected(res, err, 1)
}

func (r *Roach) getConf(key string, conf interface{}) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	q := `SELECT ` + ColValue + ` FROM ` + TblConfigurations + ` WHERE ` + ColKey + `=$1`
	var confB []byte
	if err := r.db.QueryRow(q, key).Scan(&confB); err != nil {
		if err == sql.ErrNoRows {
			return errors.NewNotFoundf("config not found")
		}
		return err
	}
	if err := json.Unmarshal(confB, conf); err != nil {
		return errors.Newf("Unmarshal config: %v", err)
	}
	return nil
}
