package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	baseURL = "http://localhost:8080"
)

type Car struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Color string `json:"color"`
}

type PagedResponse struct {
	Data       []Car `json:"data"`
	TotalItems int   `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

func main() {
	// Define command line flags
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listPage := listCmd.Int("page", 1, "Page number")
	listPageSize := listCmd.Int("page-size", 10, "Page size")
	listMake := listCmd.String("make", "", "Filter by make")
	listModel := listCmd.String("model", "", "Filter by model")
	listYear := listCmd.Int("year", 0, "Filter by year")
	listColor := listCmd.String("color", "", "Filter by color")
	listSort := listCmd.String("sort", "", "Sort field (make, model, year, color)")
	listOrder := listCmd.String("order", "asc", "Sort order (asc, desc)")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getID := getCmd.String("id", "", "Car ID to retrieve")

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createID := createCmd.String("id", "", "Car ID")
	createMake := createCmd.String("make", "", "Car make")
	createModel := createCmd.String("model", "", "Car model")
	createYear := createCmd.Int("year", 0, "Car year")
	createColor := createCmd.String("color", "", "Car color")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateID := updateCmd.String("id", "", "Car ID to update")
	updateMake := updateCmd.String("make", "", "Car make")
	updateModel := updateCmd.String("model", "", "Car model")
	updateYear := updateCmd.Int("year", 0, "Car year")
	updateColor := updateCmd.String("color", "", "Car color")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteCmd.String("id", "", "Car ID to delete")

	healthCmd := flag.NewFlagSet("health", flag.ExitOnError)

	// Check if a command was provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse the command
	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		listCars(*listPage, *listPageSize, *listMake, *listModel, *listYear, *listColor, *listSort, *listOrder)
	case "get":
		getCmd.Parse(os.Args[2:])
		if *getID == "" {
			fmt.Println("Error: id is required")
			getCmd.PrintDefaults()
			os.Exit(1)
		}
		getCar(*getID)
	case "create":
		createCmd.Parse(os.Args[2:])
		if *createMake == "" || *createModel == "" || *createYear <= 0 || *createColor == "" {
			fmt.Println("Error: make, model, year, and color are required")
			createCmd.PrintDefaults()
			os.Exit(1)
		}
		createCar(*createID, *createMake, *createModel, *createYear, *createColor)
	case "update":
		updateCmd.Parse(os.Args[2:])
		if *updateID == "" {
			fmt.Println("Error: id is required")
			updateCmd.PrintDefaults()
			os.Exit(1)
		}
		updateCar(*updateID, *updateMake, *updateModel, *updateYear, *updateColor)
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteID == "" {
			fmt.Println("Error: id is required")
			deleteCmd.PrintDefaults()
			os.Exit(1)
		}
		deleteCar(*deleteID)
	case "health":
		healthCmd.Parse(os.Args[2:])
		checkHealth()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("CarFlow CLI - A command-line interface for the CarFlow API")
	fmt.Println("\nUsage:")
	fmt.Println("  carflow-cli [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  list    - List all cars with optional filtering and pagination")
	fmt.Println("  get     - Get a specific car by ID")
	fmt.Println("  create  - Create a new car")
	fmt.Println("  update  - Update an existing car")
	fmt.Println("  delete  - Delete a car")
	fmt.Println("  health  - Check API health")
	fmt.Println("  help    - Show this help message")
	fmt.Println("\nRun 'carflow-cli [command] -h' for more information on a command.")
}

func listCars(page, pageSize int, make, model string, year int, color, sort, order string) {
	// Build query parameters
	queryParams := []string{}

	if page > 0 {
		queryParams = append(queryParams, fmt.Sprintf("page=%d", page))
	}

	if pageSize > 0 {
		queryParams = append(queryParams, fmt.Sprintf("page_size=%d", pageSize))
	}

	if make != "" {
		queryParams = append(queryParams, fmt.Sprintf("make=%s", make))
	}

	if model != "" {
		queryParams = append(queryParams, fmt.Sprintf("model=%s", model))
	}

	if year > 0 {
		queryParams = append(queryParams, fmt.Sprintf("year=%d", year))
	}

	if color != "" {
		queryParams = append(queryParams, fmt.Sprintf("color=%s", color))
	}

	if sort != "" {
		queryParams = append(queryParams, fmt.Sprintf("sort=%s", sort))
	}

	if order != "" {
		queryParams = append(queryParams, fmt.Sprintf("order=%s", order))
	}

	// Build URL
	url := fmt.Sprintf("%s/cars", baseURL)
	if len(queryParams) > 0 {
		url += "?" + strings.Join(queryParams, "&")
	}

	// Send request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching cars: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: %s", string(body))
	}

	// Parse response
	var pagedResponse PagedResponse
	if err := json.Unmarshal(body, &pagedResponse); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Print cars
	fmt.Printf("Page %d of %d (Total items: %d)\n\n",
		pagedResponse.Page,
		pagedResponse.TotalPages,
		pagedResponse.TotalItems,
	)

	if len(pagedResponse.Data) == 0 {
		fmt.Println("No cars found.")
		return
	}

	for _, car := range pagedResponse.Data {
		fmt.Printf("ID: %s\n", car.ID)
		fmt.Printf("Make: %s\n", car.Make)
		fmt.Printf("Model: %s\n", car.Model)
		fmt.Printf("Year: %d\n", car.Year)
		fmt.Printf("Color: %s\n", car.Color)
		fmt.Println("----------")
	}
}

func getCar(id string) {
	url := fmt.Sprintf("%s/cars/%s", baseURL, id)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching car: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var car Car
	if err := json.Unmarshal(body, &car); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Printf("ID: %s\n", car.ID)
	fmt.Printf("Make: %s\n", car.Make)
	fmt.Printf("Model: %s\n", car.Model)
	fmt.Printf("Year: %d\n", car.Year)
	fmt.Printf("Color: %s\n", car.Color)
}

func createCar(id, make, model string, year int, color string) {
	car := Car{
		ID:    id,
		Make:  make,
		Model: model,
		Year:  year,
		Color: color,
	}

	payload, err := json.Marshal(car)
	if err != nil {
		log.Fatalf("Error creating payload: %v", err)
	}

	url := fmt.Sprintf("%s/cars", baseURL)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		log.Fatalf("Error creating car: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var createdCar Car
	if err := json.Unmarshal(body, &createdCar); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Println("Car created successfully:")
	fmt.Printf("ID: %s\n", createdCar.ID)
	fmt.Printf("Make: %s\n", createdCar.Make)
	fmt.Printf("Model: %s\n", createdCar.Model)
	fmt.Printf("Year: %d\n", createdCar.Year)
	fmt.Printf("Color: %s\n", createdCar.Color)
}

func updateCar(id, make, model string, year int, color string) {
	// First get the existing car
	url := fmt.Sprintf("%s/cars/%s", baseURL, id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching car: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		resp.Body.Close()
		os.Exit(1)
	}

	var existingCar Car
	if err := json.NewDecoder(resp.Body).Decode(&existingCar); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}
	resp.Body.Close()

	// Update fields if provided
	if make != "" {
		existingCar.Make = make
	}

	if model != "" {
		existingCar.Model = model
	}

	if year > 0 {
		existingCar.Year = year
	}

	if color != "" {
		existingCar.Color = color
	}

	// Send update request
	payload, err := json.Marshal(existingCar)
	if err != nil {
		log.Fatalf("Error creating payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(payload)))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Error updating car: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var updatedCar Car
	if err := json.Unmarshal(body, &updatedCar); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Println("Car updated successfully:")
	fmt.Printf("ID: %s\n", updatedCar.ID)
	fmt.Printf("Make: %s\n", updatedCar.Make)
	fmt.Printf("Model: %s\n", updatedCar.Model)
	fmt.Printf("Year: %d\n", updatedCar.Year)
	fmt.Printf("Color: %s\n", updatedCar.Color)
}

func deleteCar(id string) {
	url := fmt.Sprintf("%s/cars/%s", baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error deleting car: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Printf("Car with ID '%s' has been deleted successfully.\n", id)
}

func checkHealth() {
	url := fmt.Sprintf("%s/healthz", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error checking health: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API health check failed: %s\n", string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	fmt.Println("API Health Status:")
	for k, v := range result {
		fmt.Printf("%s: %v\n", k, v)
	}
}
