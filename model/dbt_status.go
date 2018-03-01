package model

import "time"

type DBTStatus struct {
	ObfuscatedAddress string
	ExpiresAt         time.Time
}

func (s DBTStatus) HasValue() bool {
	return !s.ExpiresAt.IsZero()
}
