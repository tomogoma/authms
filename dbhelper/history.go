package dbhelper

import (
	"fmt"

	"database/sql"
	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
)

func (m *DBHelper) SaveHistory(h *authms.History) error {
	if err := validateHistory(h); err != nil {
		return err
	}
	if err := m.InitDBConnIfNotInitted(); err != nil {
		return err
	}
	q := `
	INSERT INTO history (userID, accessMethod, successful, devID, ipAddress, date)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP())
		 RETURNING id
	`
	err := m.db.QueryRow(q, h.UserID, h.AccessType, h.SuccessStatus,
		h.DevID, h.IpAddress).Scan(&h.ID)
	if err != nil {
		return errors.Newf("error inserting history: %v", err)
	}
	return nil
}

func (m *DBHelper) GetHistory(userID int64, offset, count int, acMs ...string) ([]*authms.History, error) {
	if err := m.InitDBConnIfNotInitted(); err != nil {
		return nil, err
	}
	acMFilter := ""
	for i, acM := range acMs {
		if i == 0 {
			acMFilter = fmt.Sprintf("AND (accessMethod = '%s'", acM)
			continue
		}
		acMFilter = fmt.Sprintf("%s OR accessMethod = '%s'", acMFilter, acM)
	}
	if acMFilter != "" {
		acMFilter += ")"
	}
	q := fmt.Sprintf(`
		SELECT id, accessMethod, successful, userID, date, devID, ipAddress
		FROM history
		WHERE userID = $1 %s
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, acMFilter)
	r, err := m.db.Query(q, userID, count, offset)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	hists := make([]*authms.History, 0)
	for r.Next() {
		var devID, ipAddr sql.NullString
		d := &authms.History{}
		err = r.Scan(&d.ID, &d.AccessType, &d.SuccessStatus, &d.UserID,
			&d.Date, &devID, &ipAddr)
		d.DevID = devID.String
		d.IpAddress = ipAddr.String
		if err != nil {
			return nil, errors.Newf("error scanning result row: %v",
				err)
		}
		hists = append(hists, d)
	}
	if err := r.Err(); err != nil {
		return nil, errors.Newf("error iterating resultset: %v", err)
	}
	return hists, nil
}

func validateHistory(h *authms.History) error {
	if h.UserID < 1 {
		return errors.New("userID was invalid")
	}
	if h.AccessType == "" {
		return errors.New("access type was empty")
	}
	return nil
}
