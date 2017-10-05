package db

import (
	"database/sql"
	"reflect"

	"github.com/tomogoma/authms/model"
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
