package model

import "time"

const (
	RoleCustomer = "customer"
	RoleVendor   = "vendor"
	RoleAdmin    = "admin"
)

type User struct {
	ID         int64     `json:"id"`
	Identifier string    `json:"identifier"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}

type SetRoleRequest struct {
	Role string `json:"role"`
}
