package cockroach

import "fmt"

type DSN struct {
	UName       string `yaml:"userName,omitempty"`
	Password    string `yaml:"password,omitempty"`
	Host        string `yaml:"host,omitempty"`
	DB          string `yaml:"dbName,omitempty"`
	SslCert     string `yaml:"sslCert,omitempty"`
	SslKey      string `yaml:"sslKey,omitempty"`
	SslRootCert string `yaml:"sslRootCert,omitempty"`
}

func (d DSN) DBName() string {
	return d.DB
}

func (d DSN) Validate() error {
	return nil
}

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
