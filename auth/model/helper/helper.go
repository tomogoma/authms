package helper

import "errors"

var ErrorNilDB = errors.New("db cannot be nil")

type Model interface {
	TableName() string
	TableDesc() string
}
