package model

import "time"

type User struct {
	ID         string
	Type       UserType
	UserName   Username
	Phone      VerifLogin
	Email      VerifLogin
	Facebook   Facebook
	Groups     []Group
	CreateDate time.Time
	UpdateDate time.Time
	Devices    []Device
	JWT        string
}
