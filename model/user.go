package model

import (
	"time"
	"strings"
	"strconv"
	errors "github.com/tomogoma/go-typed-errors"
)

type User struct {
	ID         string
	JWT        string
	Type       UserType
	UserName   Username
	Phone      VerifLogin
	Email      VerifLogin
	Facebook   Facebook
	Groups     []Group
	Devices    []Device
	CreateDate time.Time
	UpdateDate time.Time
}

func (u User) HasValue() bool {
	return u.ID != ""
}

type NumericQuery struct {
	CheckVal float64
	IsGT     bool
	IsLT     bool
	IsEq     bool
}

type UsersQuery struct {
	AccessLevelsIn []string
	ProcessedACLs  []NumericQuery
	GroupNamesIn   []string
	MatchAll       bool
	MatchAllACLs   bool
}

func (uq *UsersQuery) Process() error {

	uq.ProcessedACLs = make([]NumericQuery, 0)

	for i, acl := range uq.AccessLevelsIn {

		if acl == "" {
			continue
		}

		parts := strings.Split(acl, "_")

		if len(parts) == 1 {
			val, err := strconv.ParseFloat(parts[0], 10)
			if err != nil {
				return errors.NewClientf("invalid access level filter at index %d: %v", i, err)
			}

			uq.ProcessedACLs = append(uq.ProcessedACLs, NumericQuery{
				CheckVal: val, IsEq: true,
			})
			continue
		}

		val, err := strconv.ParseFloat(parts[1], 10)
		if err != nil {
			return errors.NewClientf("invalid access level filter at index %d: %v", i, err)
		}

		nq := NumericQuery{CheckVal: val}
		switch parts[0] {
		case "gt":
			nq.IsGT = true
		case "lt":
			nq.IsLT = true
		case "gteq":
			nq.IsGT = true
			nq.IsEq = true
		case "lteq":
			nq.IsLT = true
			nq.IsEq = true
		default:
			return errors.NewClientf("invalid access level filter at index %d:"+
				"the comparator is invalid", i)
		}
		uq.ProcessedACLs = append(uq.ProcessedACLs, nq)
	}

	return nil
}
