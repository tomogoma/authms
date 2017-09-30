package model

import "time"

type User struct {
	ID         string
	Type       UserType
	UserName   Username
	Phone      Phone
	Email      Email
	Facebook   Facebook
	Groups     []Group
	CreateDate time.Time
	UpdateDate time.Time
}
