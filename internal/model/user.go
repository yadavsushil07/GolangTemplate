package model

import "time"

const (
	RoleCustomer = "customer"
	RoleVendor   = "vendor"
)

type User struct {
	ID         int64     `json:"id"`
	Identifier string    `json:"identifier"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}
