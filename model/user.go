package model

import "time"

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

type UsersQuery struct {
	AccessLevelsIn []string
	GroupNamesIn   []string
}
