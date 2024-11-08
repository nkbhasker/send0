package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	uid "github.com/usesend0/send0/internal/uid"
)

const (
	IdentityProviderLocal  IdentityProvider = "LOCAL"
	IdentityProviderGoogle IdentityProvider = "GOOGLE"
	IdentityProviderGithub IdentityProvider = "GITHUB"
)

var IdentityProviderTypeCreateQuery = fmt.Sprintf(
	`CREATE TYPE %s AS ENUM ('%s','%s','%s');`,
	DBTypeIdentityProvider,
	IdentityProviderLocal,
	IdentityProviderGoogle,
	IdentityProviderGithub,
)

var _ sql.Scanner = (*IdentityProvider)(nil)
var _ driver.Valuer = (*IdentityProvider)(nil)

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindById(ctx context.Context, id uid.UID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	// WithTx(tx pgx.Tx) UserRepository
}

type IdentityProvider string

type User struct {
	Base
	FirstName        string           `json:"firstName" db:"first_name"`
	LastName         string           `json:"lastName" db:"last_name"`
	Email            string           `json:"email" db:"email" gorm:"not null;index:idx_email,unique,where:email IS NOT NULL" `
	EmailVerified    bool             `json:"emailVerified" db:"email_verified"`
	IdentityProvider IdentityProvider `json:"identityProvider" db:"identity_provider" gorm:"not null;type:identity_provider;default:'LOCAL'"`
}

type userRepository struct {
	*baseRepository
}

func NewUserRepository(baseRepository *baseRepository) UserRepository {
	return &userRepository{
		baseRepository,
	}
}

// func (u *userRepository) WithTx(tx pgx.Tx) UserRepository {
// 	return NewUserRepository(u.baseRepository.WithTx(tx))
// }

func (u *userRepository) Save(ctx context.Context, user *User) error {
	stmt := `INSERT INTO users (
		id, 
		first_name, 
		last_name, 
		email, 
		email_verified, 
		identity_provider
	) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	_, err := u.DB.Connection().Exec(ctx, stmt, u.UID(user.Id), user.FirstName, user.LastName, user.Email, user.EmailVerified, user.IdentityProvider)

	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	stmt := `SELECT * FROM users WHERE email = $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, email).Scan(&user)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func (r *userRepository) FindById(ctx context.Context, id uid.UID) (*User, error) {
	var user User
	stmt := `SELECT * FROM users WHERE id = $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(&user)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func (r *userRepository) Update(ctx context.Context, user *User) error {
	stmt := `UPDATE users SET first_name = $1, last_name = $2 WHERE id = $3`
	_, err := r.DB.Connection().Exec(ctx, stmt, user.FirstName, user.LastName, user.Id)

	return err
}

func (i *IdentityProvider) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*i = IdentityProvider(v)
	default:
		return fmt.Errorf("invalid identity provider: %v", value)
	}

	return nil
}

func (i IdentityProvider) Value() (driver.Value, error) {
	if i == "" {
		return nil, nil
	}

	return string(i), nil
}
