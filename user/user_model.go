package user

import (
	"time"

	"github.com/kamva/mgm/v3"
)

type User struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string `json:"name" bson:"name" validate:"required"`
	Age              int    `json:"age" bson:"age" validate:"required"`
}

func (u *User) Creating() error {
	u.CreatedAt = time.Now().UTC().Truncate(time.Second)
	u.UpdatedAt = u.CreatedAt
	return nil
}

func (u *User) Saving() error {
	u.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	return nil
}
