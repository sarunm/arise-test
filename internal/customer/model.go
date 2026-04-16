package customer

import (
	"errors"
	"time"
)

type Customer struct {
	ID        int    `gorm:"primaryKey"`
	FirstName string `gorm:"size:100;not null"`
	LastName  string `gorm:"size:100;not null"`
	Email     string `gorm:"size:255;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type CreateCustomerRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type UpdateCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type CustomerResponse struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Customer) ToResponse() CustomerResponse {
	return CustomerResponse{
		ID:        c.ID,
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Email:     c.Email,
		CreatedAt: c.CreatedAt,
	}
}
