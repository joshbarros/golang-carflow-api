package car

// Car represents a car entity in the system
type Car struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Color string `json:"color"`
}
