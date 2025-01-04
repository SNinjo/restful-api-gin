package user

import "time"

type EditUserDto struct {
	Name string `json:"name" example:"Jo Liao"`
	Age  int    `json:"age" example:"22"`
}

type UserDto struct {
	EditUserDto
	ID        string    `json:"id" example:"63e57e2e1b2e4d0f9c564b33"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T00:00:000Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-01T00:00:000Z"`
}
