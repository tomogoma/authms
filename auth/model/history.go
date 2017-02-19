package model

import (
	"fmt"

	"github.com/tomogoma/authms/proto/authms"
	"github.com/tomogoma/go-commons/errors"
)

const (
	AccessLogin = "LOGIN"
	AccessRegistration = "REGISTER"
	AccessUpdate = "UPDATE"
	AccessPassChange = "PASSWORD_CHANGE"
)

var accessTypes = []string{AccessLogin, AccessRegistration, AccessUpdate, AccessPassChange}

func (m *Model) SaveHistory(h *authms.History) error {
	if err := validateHistory(h); err != nil {
		return err
	}
	q := `
	INSERT INTO history (userID, accessMethod, successful, devID, ipAddress, date)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP())
		 RETURNING id
	`
	return m.db.QueryRow(q, h.UserID, h.AccessType, h.SuccessStatus,
		h.DevID, h.IpAddress).Scan(&h.ID)
}

func (m *Model) GetHistory(userID int64, offset, count int, acMs ...string) ([]*authms.History, error) {
	acMFilter := ""
	for i, acM := range acMs {
		if err := in(acM, accessTypes); err != nil {
			return nil, errors.Newf("access type not valid: %v", err)
		}
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
		d := &authms.History{}
		err = r.Scan(&d.ID, &d.AccessType, &d.SuccessStatus, &d.UserID,
			&d.Date, &d.DevID, &d.IpAddress)
		if err != nil {
			return nil, err
		}
		hists = append(hists, d)
	}
	return hists, r.Err()
}

func validateHistory(h *authms.History) error {
	if h.UserID < 1 {
		return errors.New("userID was invalid")
	}
	if err := in(h.AccessType, accessTypes); err != nil {
		return errors.Newf("access type not valid: %v", err)
	}
	return nil
}

func in(check string, valids []string) error {
	for _, valid := range valids {
		if check == valid {
			return nil
		}
	}
	return errors.Newf("%s not found in %v", check, valids)
}
