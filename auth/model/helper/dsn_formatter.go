package helper

import "fmt"

type DSNFormatter interface {
	FormatDSN() string
}

type DSN struct {
	UName    string
	Password string
	Host     string
	DB       string
}

func (d DSN) FormatDSN() string {

	if d.UName != "" {
		if d.Password != "" {
			return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=verify-full",
				d.UName, d.Password, d.Host, d.DB,
			)
		}
		return fmt.Sprintf("postgres://%s@%s/%s?sslmode=verify-full",
			d.UName, d.Host, d.DB,
		)
	}

	return fmt.Sprintf("postgres://%s/%s?sslmode=verify-full", d.Host, d.DB)
}
