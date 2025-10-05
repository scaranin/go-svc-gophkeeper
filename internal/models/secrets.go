package models

import (
	"time"
)

type SecretType string

const (
	TypeUnspecified SecretType = "unspecified"
	TypeLogin       SecretType = "login"
	TypeCard        SecretType = "card"
	TypeText        SecretType = "text"
	TypeBinary      SecretType = "binary"
)

type Secret struct {
	ID            int        `db:"id"`
	UserID        int        `db:"user_id"`
	Type          SecretType `db:"type"`
	Name          string     `db:"name"`
	Metadata      []byte     `db:"metadata"`
	EncryptedData []byte     `db:"encrypted_data"`
	Version       int        `db:"version"`
	DeletedAt     *time.Time `db:"deleted_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}
