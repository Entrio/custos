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
		ID uuid.UUID `json:"id" gorm:"primaryKey,<-:create"`
	}

	DT struct {
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at,omitempty"`
	}

	User struct {
		Base
		Email         string     `json:"email"`
		Position      string     `json:"position"`
		FirstName     string     `json:"first_name"`
		LastName      string     `json:"last_name"`
		Verified      bool       `json:"verified"`
		Enabled       bool       `json:"enabled"`
		Protected     bool       `json:"protected" gorm:"->"`
		DisableReason *string    `json:"disable_reason,omitempty" gorm:"column:enabled_reason"`
		DisabledDate  *time.Time `json:"disabled_date,omitempty" gorm:"column:enabled_date"`
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
		Name          string  `json:"name"`
		Description   *string `json:"description"`
		ParentGroupID *string `json:"-" gorm:"column:parent_group"`
		ParentGroup   *Group  `json:"parent_group,omitempty"`
		Enabled       bool    `json:"enabled"`
		Protected     bool    `json:"protected" gorm:"<-:false"`
		Users         []*User `json:"users" gorm:"many2many:user_group"`
		DT
	}

	Service struct {
		Base
		Name        string              `json:"name"`
		Description string              `json:"description"`
		Verbs       []Verb              `json:"verbs" gorm:"many2many:service_verb;save_associations:false"`
		Enabled     bool                `json:"enabled"`
		Protected   bool                `json:"protected"`
		GroupVerbs  []ServiceGroupVerbs `json:"groups"`
		DT
	}
	ServiceGroupVerbs struct {
		ServiceID uuid.UUID `json:"service_id"`
		VerbID    uuid.UUID `json:"verb_id"`
		GroupID   uuid.UUID `json:"group_id"`
	}

	Verb struct {
		Base
		Name string `json:"name"`
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
