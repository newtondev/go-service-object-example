package storage

import (
	"context"

	"github.com/newtondev/service_object/pkg/entities"
	"github.com/newtondev/service_object/pkg/errors"
)

// MemStore is a memory storage for users.
type MemStore struct {
	Users []entities.User
}

// Unique checks if a email exists in the database.
func (s *MemStore) Unique(ctx context.Context, email string) error {
	for _, u := range s.Users {
		if u.Email == email {
			return errors.ErrEmailExists
		}
	}

	return nil
}

// Create creates user in the database for a form.
func (s *MemStore) Create(ctx context.Context, f *entities.Form) (*entities.User, error) {
	u := entities.User{
		ID:       len(s.Users) + 1,
		Password: f.Password,
		Email:    f.Email,
	}

	s.Users = append(s.Users, u)

	return &u, nil
}
