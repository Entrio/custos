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
		ID uuid.UUID `json:"id" gorm:"primaryKey"`
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
		Name        string  `json:"name"`
		Description *string `json:"description"`
		ParentGroup *Group  `json:"parent_group"`
		DT
	}

	Service struct {
		Base
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	Model struct {
		Base
		Name string `json:"name"`
	}

	/**
	Controls group access to a given service
	*/
	GroupService struct {
		Base
		Group   Group   `json:"group"`
		Service Service `json:"service"`
		Verb    string  `json:"verb"`
		Action  string  `json:"action"`
	}

	GroupServiceModel struct {
	}
)
