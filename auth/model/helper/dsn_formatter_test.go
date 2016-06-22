package helper_test

import (
	"testing"

	"fmt"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
)

func TestDSN_FormatDSN(t *testing.T) {

	sslcert := "/path/to/ssl/node.cert"
	sslkey := "/path/to/ssl/node.key"
	sslrootcert := "/path/to/ca.cert"
	host := "host:26257"
	pass := "p@$$"
	uname := "uname"

	type tcase struct {
		dsn  helper.DSN
		desc string
		exp  string
	}

	cases := []tcase{
		tcase{
			dsn: helper.DSN{
				UName:       uname,
				Password:    pass,
				Host:        host,
				DB:          testhelper.DBName,
				SslCert:     sslcert,
				SslKey:      sslkey,
				SslRootCert: sslrootcert,
			},
			desc: "all fields provided",
			exp: fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
				uname, pass, host, testhelper.DBName, sslcert, sslkey, sslrootcert),
		},
		tcase{
			dsn: helper.DSN{
				Host:        host,
				DB:          testhelper.DBName,
				SslCert:     sslcert,
				SslKey:      sslkey,
				SslRootCert: sslrootcert,
			},
			desc: "no uname/password",
			exp: fmt.Sprintf("postgres://%s/%s?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
				host, testhelper.DBName, sslcert, sslkey, sslrootcert),
		},
		tcase{
			dsn: helper.DSN{
				UName:       uname,
				Host:        host,
				DB:          testhelper.DBName,
				SslCert:     sslcert,
				SslKey:      sslkey,
				SslRootCert: sslrootcert,
			},
			desc: "no password",
			exp: fmt.Sprintf("postgres://%s@%s/%s?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
				uname, host, testhelper.DBName, sslcert, sslkey, sslrootcert),
		},
		tcase{
			dsn: helper.DSN{
				UName:       uname,
				Password:    pass,
				DB:          testhelper.DBName,
				SslCert:     sslcert,
				SslKey:      sslkey,
				SslRootCert: sslrootcert,
			},
			desc: "no host",
			exp: fmt.Sprintf("postgres://%s:%s@127.0.0.1:26257/%s?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
				uname, pass, testhelper.DBName, sslcert, sslkey, sslrootcert),
		},
		tcase{
			dsn: helper.DSN{
				UName:       uname,
				Password:    pass,
				Host:        host,
				SslCert:     sslcert,
				SslKey:      sslkey,
				SslRootCert: sslrootcert,
			},
			desc: "No DB",
			exp: fmt.Sprintf("postgres://%s:%s@%s/?sslmode=verify-full&sslcert=%s&sslkey=%s&sslrootcert=%s",
				uname, pass, host, sslcert, sslkey, sslrootcert),
		},
		tcase{
			dsn: helper.DSN{
				UName:    uname,
				Password: pass,
				Host:     host,
				DB:       testhelper.DBName,
			},
			desc: "no SSL",
			exp: fmt.Sprintf("postgres://%s:%s@%s/%s",
				uname, pass, host, testhelper.DBName),
		},
		tcase{
			dsn:  helper.DSN{},
			desc: "no field provided",
			exp:  "postgres://127.0.0.1:26257/",
		},
	}

	for _, c := range cases {

		act := c.dsn.FormatDSN()
		if act != c.exp {
			t.Errorf("Test %s\nExpect:\t%s\nGot\t\t%s", c.desc, c.exp, act)
		}
	}
}
