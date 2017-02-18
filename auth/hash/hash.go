package hash

import "golang.org/x/crypto/bcrypt"

type Hasher struct {
}

func (h Hasher) Hash(pass string) ([]byte, error) {
	passB := []byte(pass)
	return bcrypt.GenerateFromPassword(passB, bcrypt.DefaultCost)
}

func (h Hasher) CompareHash(pass string, passHB []byte) bool {
	passB := []byte(pass)
	err := bcrypt.CompareHashAndPassword(passHB, passB)
	return err == nil
}
