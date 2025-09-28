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

// ConvertSecretType конвертирует proto SecretType в models.SecretType
func ConvertSecretType(protoType string) SecretType {
	switch protoType {
	case "SECRET_TYPE_LOGIN":
		return TypeLogin
	case "SECRET_TYPE_CARD":
		return TypeCard
	case "SECRET_TYPE_TEXT":
		return TypeText
	case "SECRET_TYPE_BINARY":
		return TypeBinary
	default:
		return TypeUnspecified
	}
}

// ToString возвращает строковое представление для хранения в БД
func (st SecretType) ToString() string {
	return string(st)
}

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
