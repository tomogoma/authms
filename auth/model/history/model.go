package history

import (
	"database/sql"
	"fmt"

	"time"

	"github.com/tomogoma/authms/auth/model/helper"
)

const (
	dateFormat = time.RFC3339 // "2006-01-02T15:04:05Z07:00"
)

type Model struct {
	db *sql.DB
}

func NewModel(db *sql.DB) (*Model, error) {

	if db == nil {
		return nil, helper.ErrorNilDB
	}

	return &Model{db: db}, nil
}

func (m *Model) Save(ld History) (int, error) {

	if err := ld.Validate(); err != nil {
		return 0, err
	}

	qStr := `
	INSERT INTO history (userID, accessMethod, successful, date, forServiceID, ipAddress, referral)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id
	`

	var detsID int
	date := ld.date.Format(dateFormat)
	err := m.db.QueryRow(qStr, ld.userID, ld.accessMethod, ld.successful, date,
		ld.forService, ld.ipAddress, ld.referral).Scan(&detsID)
	return detsID, err
}

func (m *Model) Get(userID, offset, count int, acMs ...int) ([]*History, error) {

	acMFilter := ""

	for i, acM := range acMs {

		if err := validateAcM(acM); err != nil {
			return nil, err
		}

		if i == 0 {
			acMFilter = fmt.Sprintf("AND (accessMethod = %d", acM)
			continue
		}

		acMFilter = fmt.Sprintf("%s OR accessMethod = %d", acMFilter, acM)
	}

	if acMFilter != "" {
		acMFilter += ")"
	}

	qStr := fmt.Sprintf(`
		SELECT id, accessMethod, successful, userID, date, forServiceID, ipAddress, referral
		FROM history
		WHERE userID = $1 %s
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, acMFilter)

	r, err := m.db.Query(qStr, userID, count, offset)
	if err != nil {
		return nil, err
	}

	dets := make([]*History, 0)
	for r.Next() {

		d := &History{}
		var tmStmp string
		r.Scan(&d.id, &d.accessMethod, &d.successful, &d.userID, &tmStmp, &d.forService, &d.ipAddress, &d.referral)
		d.date, err = time.Parse(dateFormat, tmStmp)
		if err != nil {
			return nil, err
		}
		dets = append(dets, d)
	}

	return dets, r.Err()
}
