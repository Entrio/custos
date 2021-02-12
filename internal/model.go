package internal

import (
	"github.com/satori/go.uuid"
	"time"
)

type OathkeeperRequest struct {
	Subject string  `json:"subject" validate:"required,uuid4"`
	Verb    string  `json:"verb" validate:"required"`
	Service string  `json:"service" validate:"required"`
	Model   *string `json:"model" validate:"required"`
	ModelID *string `json:"id" validate:"required"`
	Action  *string `json:"action" validate:"required"`
}

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

	KratosUser struct {
		ID     string `json:"id"`
		Traits struct {
			Name struct {
				First string `json:"first"`
				Last  string `json:"last"`
			} `json:"name"`
			Email    string `json:"email"`
			Position string `json:"position"`
		} `json:"traits"`
		VerifiableAddresses []struct {
			ID         string    `json:"id"`
			Email      string    `json:"value"`
			Verified   bool      `json:"verified"`
			Status     string    `json:"status"`
			VerifiedAt time.Time `json:"verified_at"`
		} `json:"verifiable_addresses"`
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
