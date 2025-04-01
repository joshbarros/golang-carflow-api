package car

import "time"

// Car represents a car entity in the system
type Car struct {
	ID        string    `json:"id" db:"id"`
	Make      string    `json:"make" db:"make"`
	Model     string    `json:"model" db:"model"`
	Year      int       `json:"year" db:"year"`
	Color     string    `json:"color" db:"color"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
