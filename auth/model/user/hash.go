package user

import "golang.org/x/crypto/bcrypt"

type HashFunc func(pass string) ([]byte, error)
type ValidatePassFunc func(pass string, passHB []byte) bool

func Hash(pass string) ([]byte, error) {

	passB := []byte(pass)
	return bcrypt.GenerateFromPassword(passB, bcrypt.DefaultCost)
}

func CompareHash(pass string, passHB []byte) bool {

	passB := []byte(pass)
	err := bcrypt.CompareHashAndPassword(passHB, passB)
	return err == nil
}
