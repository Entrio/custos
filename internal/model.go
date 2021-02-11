package internal

import (
	"github.com/satori/go.uuid"
	"time"
)

type (
	Base struct {
		ID uuid.UUID `json:"id"`
	}

	DT struct {
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at,omitempty"`
	}

	User struct {
		Base
		Email    string `json:"email"`
		Position string `json:"position"`
		FirsName string `json:"firs_name"`
		LastName string `json:"last_name"`
		Verified bool   `json:"verified"`
		Enabled  bool   `json:"enabled"`
		DT
	}

	Group struct {
		Base
		// Base model contents
		DT
	}

	Permission struct {
		Base
		// Base model contents
		DT
	}
)
