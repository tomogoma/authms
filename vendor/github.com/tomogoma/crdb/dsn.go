package crdb

import (
	"fmt"
	"strings"
)

// More detail at
// https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
type Config struct {
	User           string `json:"user,omitempty" yaml:"user,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	Host           string `json:"host,omitempty" yaml:"host,omitempty"`
	Port           int    `json:"port,omitempty" yaml:"port,omitempty"`
	DBName         string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
	ConnectTimeout int    `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"`
	SSLMode        string `json:"sslMode,omitempty" yaml:"sslMode,omitempty"`
	SSLCert        string `json:"sslCert,omitempty" yaml:"sslCert,omitempty"`
	SSLKey         string `json:"sslKey,omitempty" yaml:"sslKey,omitempty"`
	SSLRootCert    string `json:"sslRootCert,omitempty" yaml:"sslRootCert,omitempty"`
}

// FormatDSN formats Config values into a connection as per
// https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func (d Config) FormatDSN() string {
	dsn := ""
	if d.User != "" {
		dsn = fmt.Sprintf("%suser='%s' ", dsn, d.User)
	}
	if d.Password != "" {
		dsn = fmt.Sprintf("%spassword='%s' ", dsn, d.Password)
	}
	if d.DBName != "" {
		dsn = fmt.Sprintf("%sdbname='%s' ", dsn, d.DBName)
	}
	if d.Host != "" {
		dsn = fmt.Sprintf("%shost='%s' ", dsn, d.Host)
	}
	if d.Port > 0 {
		dsn = fmt.Sprintf("%sport=%d ", dsn, d.Port)
	}
	if d.ConnectTimeout > 0 {
		dsn = fmt.Sprintf("%sconnect_timeout=%d ", dsn, d.ConnectTimeout)
	}
	if d.SSLMode != "" {
		dsn = fmt.Sprintf("%ssslmode='%s' ", dsn, d.SSLMode)
	}
	if d.SSLCert != "" {
		dsn = fmt.Sprintf("%ssslcert='%s' ", dsn, d.SSLCert)
	}
	if d.SSLKey != "" {
		dsn = fmt.Sprintf("%ssslkey='%s' ", dsn, d.SSLKey)
	}
	if d.SSLRootCert != "" {
		dsn = fmt.Sprintf("%ssslrootcert='%s' ", dsn, d.SSLRootCert)
	}
	return strings.TrimSpace(dsn)
}
