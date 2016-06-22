package helper_test

import (
	"testing"

	"strings"

	"bitbucket.org/tomogoma/auth-ms/auth/model/helper"
	"bitbucket.org/tomogoma/auth-ms/auth/model/testhelper"
)

type model struct {
	tableName string
	tableDesc string
}

func (t model) TableName() string { return t.tableName }
func (t model) TableDesc() string { return t.tableDesc }

var table1 = model{tableName: "table1", tableDesc: "id SERIAL PRIMARY KEY, col1 STRING, col2 INT"}
var table2 = model{tableName: "table2", tableDesc: "id SERIAL PRIMARY KEY, col1 TIMESTAMP UNIQUE, col2 STRING, INDEX col2_index (col2)"}
var dupTableModel = model{tableName: "table1"}

func TestSQLDB(t *testing.T) {
	db := testhelper.InstantiateDB(t)
	defer db.Close()
}

func TestSQLDB_nilDSNFormatter(t *testing.T) {

	_, err := helper.SQLDB(nil)
	if err == nil || err != helper.ErrorNilDSNFormatter {
		t.Fatalf("Expected error %s but got %v", helper.ErrorNilDSNFormatter, err)
	}
}

func TestCreateTables(t *testing.T) {

	db := testhelper.InstantiateDB(t)
	testhelper.SetUp(nil, db, t)
	defer testhelper.TearDown(db, t)

	err := helper.CreateTables(db, table1, table2)
	if err != nil {
		t.Fatalf("helper.CreateTables(): %s", err)
	}
}

func TestCreateTables_duplicateTableNames(t *testing.T) {

	db := testhelper.InstantiateDB(t)
	testhelper.SetUp(nil, db, t)
	defer testhelper.TearDown(db, t)

	err := helper.CreateTables(db, table1, table2, dupTableModel)
	if err == nil || !strings.HasPrefix(err.Error(), helper.DupTblErrPrfx) {
		t.Fatalf("Expected an error with prefix %s but got %v", helper.DupTblErrPrfx, err)
	}
}

func TestCreateTables_badTableDesc(t *testing.T) {

	db := testhelper.InstantiateDB(t)
	testhelper.SetUp(nil, db, t)
	defer testhelper.TearDown(db, t)

	badTableDescModel := model{tableName: "bad_table_desc", tableDesc: "some bad description"}

	err := helper.CreateTables(db, table1, table2, badTableDescModel)
	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}

	r, err := db.Query("SHOW TABLES")
	if err != nil {
		t.Fatalf("db.Query(): %s", err)
	}

	if r.Next() {
		t.Errorf("Rollback of table creation did not occur")
	}

	if err := r.Err(); err != nil {
		t.Errorf("Something wicked happened iterating results: %s", err)
	}
}
