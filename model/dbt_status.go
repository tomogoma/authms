package model

import "time"

type DBTStatus struct {
	ObfuscatedAddress string
	ExpiresAt         time.Time
}
