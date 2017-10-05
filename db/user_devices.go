package db

import (
	"database/sql"
	"reflect"

	"github.com/tomogoma/authms/model"
	"github.com/tomogoma/go-commons/errors"
)

func (r *Roach) InsertUserDeviceAtomic(tx *sql.Tx, userID, devID string) (*model.Device, error) {
	if tx == nil || reflect.ValueOf(tx).IsNil() {
		return nil, errorNilTx
	}
	dev := model.Device{UserID: userID, DeviceID: devID}
	insCols := ColDesc(ColUserID, ColDevID, ColUpdateDate)
	retCols := ColDesc(ColID, ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblDeviceIDs + ` (` + insCols + `)
		VALUES ($1,$2,CURRENT_TIMESTAMP)
		RETURNING ` + retCols
	err := tx.QueryRow(q, userID, devID).Scan(&dev.ID, &dev.CreateDate, &dev.UpdateDate)
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *Roach) UserDevicesByUserID(usrID string) ([]model.Device, error) {
	cols := ColDesc(ColID, ColUserID, ColDevID, ColCreateDate, ColUpdateDate)
	q := `SELECT ` + cols + ` FROM ` + TblDeviceIDs + ` WHERE ` + ColUserID + `=$1`
	rows, err := r.db.Query(q, usrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devs []model.Device
	for rows.Next() {
		dev := model.Device{}
		err := rows.Scan(&dev.ID, &dev.UserID, &dev.DeviceID, &dev.CreateDate, &dev.UpdateDate)
		if err != nil {
			return nil, errors.Newf("scan result set row: %v", err)
		}
		devs = append(devs, dev)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}
	if len(devs) == 0 {
		return nil, errors.NewNotFound("no devices found for user")
	}
	return devs, nil
}
