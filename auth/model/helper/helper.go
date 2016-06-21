package helper

import "errors"

var ErrorNilDB = errors.New("db cannot be nil")

const (
	NoResultsErrorStr = "sql: no rows in result set"
)

type Model interface {
	TableName() string
	TableDesc() string
}
