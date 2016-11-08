package helper

import "fmt"

type DSNFormatter interface {
	FormatDSN() string
	DBName() string
}

type DSN struct {
	UName       string
	Password    string
	Host        string
	DB          string
	SslCert     string
	SslKey      string
	SslRootCert string
}

func (d DSN) DBName() string {
	return d.DB
}

//func (d DSN) FormatDSNNoDB() string {
//
//	dsnPrefix := "postgres://"
//	var dsnSuffix string
//
//	if d.SslCert != "" {
//		dsnSuffix = fmt.Sprintf("?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
//			d.SslCert, d.SslKey, d.SslRootCert)
//	}
//
//	host := d.Host
//	if d.Host == "" {
//		host = "127.0.0.1:26257"
//	}
//
//	if d.UName != "" {
//		if d.Password != "" {
//			return fmt.Sprintf("%s%s:%s@%s/%s",
//				dsnPrefix, d.UName, d.Password, host, dsnSuffix,
//			)
//		}
//		return fmt.Sprintf("%s%s@%s/%s",
//			dsnPrefix, d.UName, host, dsnSuffix,
//		)
//	}
//
//	return fmt.Sprintf("%s%s/%s", dsnPrefix, host, dsnSuffix)
//}

func (d DSN) FormatDSN() string {

	dsnPrefix := "postgres://"
	var dsnSuffix string

	if d.SslCert != "" {
		dsnSuffix = fmt.Sprintf("?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
			d.SslCert, d.SslKey, d.SslRootCert)
	}

	host := d.Host
	if d.Host == "" {
		host = "127.0.0.1:26257"
	}

	if d.UName != "" {
		if d.Password != "" {
			return fmt.Sprintf("%s%s:%s@%s/%s%s",
				dsnPrefix, d.UName, d.Password, host, d.DB, dsnSuffix,
			)
		}
		return fmt.Sprintf("%s%s@%s/%s%s",
			dsnPrefix, d.UName, host, d.DB, dsnSuffix,
		)
	}

	return fmt.Sprintf("%s%s/%s%s", dsnPrefix, host, d.DB, dsnSuffix)

}
