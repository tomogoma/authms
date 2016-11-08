package helper_test

import (
	"testing"

	"github.com/tomogoma/authms/auth/model/helper"
	"github.com/tomogoma/authms/auth/model/testhelper"
)

func TestSQLDB(t *testing.T) {
	db := testhelper.SQLDB(t)
	defer testhelper.TearDown(db, t)
}

func TestSQLDB_nilDSNFormatter(t *testing.T) {

	_, err := helper.SQLDB(nil)
	if err == nil || err != helper.ErrorNilDSNFormatter {
		t.Fatalf("Expected error %s but got %v", helper.ErrorNilDSNFormatter, err)
	}
}
