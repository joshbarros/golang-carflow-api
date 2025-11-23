package car

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Car represents a car entity in the system
type Car struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Make        string    `json:"make" db:"make"`
	Model       string    `json:"model" db:"model"`
	Year        int       `json:"year" db:"year"`
	Color       string    `json:"color,omitempty" db:"color"`
	VIN         string    `json:"vin,omitempty" db:"vin"`
	StockNumber string    `json:"stock_number,omitempty" db:"stock_number"`
	Price       *float64  `json:"price,omitempty" db:"price"`
	Mileage     *int      `json:"mileage,omitempty" db:"mileage"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description,omitempty" db:"description"`
	Images      JSONArray `json:"images,omitempty" db:"images"`
	Metadata    JSONB     `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   *string   `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy   *string   `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// JSONB is a custom type for PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// JSONArray is a custom type for PostgreSQL JSON arrays
type JSONArray []string

// Value implements the driver.Valuer interface for JSONArray
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONArray
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
