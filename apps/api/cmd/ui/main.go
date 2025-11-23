package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

const (
	apiBaseURL = "http://localhost:8080"
)

// Car represents a car entity in the system
type Car struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Color string `json:"color"`
}

// PagedResponse represents a paginated response
type PagedResponse struct {
	Data       []Car `json:"data"`
	TotalItems int   `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

// PageData holds data for rendering pages
type PageData struct {
	Title       string
	Cars        []Car
	Car         Car
	Error       string
	Message     string
	CurrentPage int
	TotalPages  int
	TotalItems  int
	PageSize    int
	Makes       []string
	Colors      []string
	Years       []int
	FilterMake  string
	FilterColor string
	FilterYear  int
	SortField   string
	SortOrder   string
}

// Define template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"sequence": func(start, end int) []int {
		var seq []int
		for i := start; i <= end; i++ {
			seq = append(seq, i)
		}
		return seq
	},
}

func main() {
	// Parse command line arguments
	port := flag.Int("port", 3000, "Port to serve the UI on")
	flag.Parse()

	// Set up templates
	templateDir := "cmd/ui/templates"
	templates := template.Must(template.New("").Funcs(templateFuncs).ParseGlob(filepath.Join(templateDir, "*.html")))

	// Set up static file server
	fs := http.FileServer(http.Dir(filepath.Join(templateDir, "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleHomePage(w, r, templates)
	})
	http.HandleFunc("/cars", func(w http.ResponseWriter, r *http.Request) {
		handleListCars(w, r, templates)
	})
	http.HandleFunc("/cars/new", func(w http.ResponseWriter, r *http.Request) {
		handleNewCar(w, r, templates)
	})
	http.HandleFunc("/cars/view/", func(w http.ResponseWriter, r *http.Request) {
		handleViewCar(w, r, templates)
	})
	http.HandleFunc("/cars/edit/", func(w http.ResponseWriter, r *http.Request) {
		handleEditCar(w, r, templates)
	})
	http.HandleFunc("/cars/delete/", func(w http.ResponseWriter, r *http.Request) {
		handleDeleteCar(w, r, templates)
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting CarFlow UI server on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// handleHomePage renders the home page
func handleHomePage(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get API health status
	healthData, err := getAPIHealth()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking API health: %v", err), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:   "CarFlow - Home",
		Message: fmt.Sprintf("API Status: %s, Uptime: %s", healthData["status"], healthData["uptime"]),
	}

	if err := templates.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleListCars handles listing cars with filtering, sorting, and pagination
func handleListCars(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	page := 1
	pageSize := 10
	make := ""
	color := ""
	year := 0
	sort := ""
	order := "asc"

	// Parse query parameters
	if r.Method == http.MethodGet {
		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
				page = p
			}
		}

		if pageSizeParam := r.URL.Query().Get("page_size"); pageSizeParam != "" {
			if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 {
				pageSize = ps
			}
		}

		make = r.URL.Query().Get("make")
		color = r.URL.Query().Get("color")

		if yearParam := r.URL.Query().Get("year"); yearParam != "" {
			if y, err := strconv.Atoi(yearParam); err == nil && y > 0 {
				year = y
			}
		}

		sort = r.URL.Query().Get("sort")
		if orderParam := r.URL.Query().Get("order"); orderParam != "" {
			order = orderParam
		}
	}

	// Fetch cars from API
	cars, totalItems, totalPages, err := getCars(page, pageSize, make, color, year, sort, order)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching cars: %v", err), http.StatusInternalServerError)
		return
	}

	// Get unique makes, colors, and years for filtering options
	makes, colors, years := getFilterOptions(cars)

	data := PageData{
		Title:       "CarFlow - Cars",
		Cars:        cars,
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		PageSize:    pageSize,
		Makes:       makes,
		Colors:      colors,
		Years:       years,
		FilterMake:  make,
		FilterColor: color,
		FilterYear:  year,
		SortField:   sort,
		SortOrder:   order,
	}

	if err := templates.ExecuteTemplate(w, "list.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleViewCar handles viewing a single car
func handleViewCar(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	id := r.URL.Path[len("/cars/view/"):]
	if id == "" {
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}

	car, err := getCar(id)
	if err != nil {
		data := PageData{
			Title: "CarFlow - Error",
			Error: fmt.Sprintf("Error fetching car: %v", err),
		}
		if err := templates.ExecuteTemplate(w, "error.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := PageData{
		Title: fmt.Sprintf("CarFlow - %s %s", car.Make, car.Model),
		Car:   car,
	}

	if err := templates.ExecuteTemplate(w, "view.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNewCar handles creating a new car
func handleNewCar(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	data := PageData{
		Title: "CarFlow - New Car",
	}

	if r.Method == http.MethodPost {
		// Parse form
		if err := r.ParseForm(); err != nil {
			data.Error = fmt.Sprintf("Error parsing form: %v", err)
			templates.ExecuteTemplate(w, "new.html", data)
			return
		}

		// Get form values
		id := r.FormValue("id")
		make := r.FormValue("make")
		model := r.FormValue("model")
		yearStr := r.FormValue("year")
		color := r.FormValue("color")

		// Validate form values
		if make == "" || model == "" || yearStr == "" || color == "" {
			data.Error = "All fields are required"
			templates.ExecuteTemplate(w, "new.html", data)
			return
		}

		// Parse year
		year, err := strconv.Atoi(yearStr)
		if err != nil || year <= 0 {
			data.Error = "Year must be a valid number"
			templates.ExecuteTemplate(w, "new.html", data)
			return
		}

		// Create car
		car := Car{
			ID:    id,
			Make:  make,
			Model: model,
			Year:  year,
			Color: color,
		}

		if err := createCar(car); err != nil {
			data.Error = fmt.Sprintf("Error creating car: %v", err)
			templates.ExecuteTemplate(w, "new.html", data)
			return
		}

		// Redirect to cars list
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}

	if err := templates.ExecuteTemplate(w, "new.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleEditCar handles editing a car
func handleEditCar(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	id := r.URL.Path[len("/cars/edit/"):]
	if id == "" {
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Parse form
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
			return
		}

		// Get form values
		make := r.FormValue("make")
		model := r.FormValue("model")
		yearStr := r.FormValue("year")
		color := r.FormValue("color")

		// Validate form values
		if make == "" || model == "" || yearStr == "" || color == "" {
			data := PageData{
				Title: "CarFlow - Edit Car",
				Error: "All fields are required",
			}
			templates.ExecuteTemplate(w, "edit.html", data)
			return
		}

		// Parse year
		year, err := strconv.Atoi(yearStr)
		if err != nil || year <= 0 {
			data := PageData{
				Title: "CarFlow - Edit Car",
				Error: "Year must be a valid number",
			}
			templates.ExecuteTemplate(w, "edit.html", data)
			return
		}

		// Update car
		car := Car{
			ID:    id,
			Make:  make,
			Model: model,
			Year:  year,
			Color: color,
		}

		if err := updateCar(car); err != nil {
			data := PageData{
				Title: "CarFlow - Edit Car",
				Error: fmt.Sprintf("Error updating car: %v", err),
			}
			templates.ExecuteTemplate(w, "edit.html", data)
			return
		}

		// Redirect to car view
		http.Redirect(w, r, "/cars/view/"+id, http.StatusSeeOther)
		return
	}

	// Get car for editing
	car, err := getCar(id)
	if err != nil {
		data := PageData{
			Title: "CarFlow - Error",
			Error: fmt.Sprintf("Error fetching car: %v", err),
		}
		if err := templates.ExecuteTemplate(w, "error.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := PageData{
		Title: fmt.Sprintf("CarFlow - Edit %s %s", car.Make, car.Model),
		Car:   car,
	}

	if err := templates.ExecuteTemplate(w, "edit.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleDeleteCar handles deleting a car
func handleDeleteCar(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	id := r.URL.Path[len("/cars/delete/"):]
	if id == "" {
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Delete car
		if err := deleteCar(id); err != nil {
			data := PageData{
				Title: "CarFlow - Error",
				Error: fmt.Sprintf("Error deleting car: %v", err),
			}
			if err := templates.ExecuteTemplate(w, "error.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Redirect to cars list
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}

	// Get car for confirmation
	car, err := getCar(id)
	if err != nil {
		data := PageData{
			Title: "CarFlow - Error",
			Error: fmt.Sprintf("Error fetching car: %v", err),
		}
		if err := templates.ExecuteTemplate(w, "error.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data := PageData{
		Title: fmt.Sprintf("CarFlow - Delete %s %s", car.Make, car.Model),
		Car:   car,
	}

	if err := templates.ExecuteTemplate(w, "delete.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// API client functions

// getFilterOptions extracts unique makes, colors, and years from cars for filter dropdowns
func getFilterOptions(cars []Car) ([]string, []string, []int) {
	makesMap := make(map[string]bool)
	colorsMap := make(map[string]bool)
	yearsMap := make(map[int]bool)

	for _, car := range cars {
		makesMap[car.Make] = true
		colorsMap[car.Color] = true
		yearsMap[car.Year] = true
	}

	makes := make([]string, 0, len(makesMap))
	for make := range makesMap {
		makes = append(makes, make)
	}

	colors := make([]string, 0, len(colorsMap))
	for color := range colorsMap {
		colors = append(colors, color)
	}

	years := make([]int, 0, len(yearsMap))
	for year := range yearsMap {
		years = append(years, year)
	}

	return makes, colors, years
}

// getAPIHealth checks the health of the API
func getAPIHealth() (map[string]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/healthz", apiBaseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API health check failed with status %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// getCars fetches cars from the API with filtering and pagination
func getCars(page, pageSize int, make, color string, year int, sort, order string) ([]Car, int, int, error) {
	// Build URL with query parameters
	url := fmt.Sprintf("%s/cars?page=%d&page_size=%d", apiBaseURL, page, pageSize)

	if make != "" {
		url += fmt.Sprintf("&make=%s", make)
	}

	if color != "" {
		url += fmt.Sprintf("&color=%s", color)
	}

	if year > 0 {
		url += fmt.Sprintf("&year=%d", year)
	}

	if sort != "" {
		url += fmt.Sprintf("&sort=%s", sort)
	}

	if order != "" {
		url += fmt.Sprintf("&order=%s", order)
	}

	// Send request
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var pagedResponse PagedResponse
	if err := json.NewDecoder(resp.Body).Decode(&pagedResponse); err != nil {
		return nil, 0, 0, err
	}

	return pagedResponse.Data, pagedResponse.TotalItems, pagedResponse.TotalPages, nil
}

// getCar fetches a single car from the API
func getCar(id string) (Car, error) {
	resp, err := http.Get(fmt.Sprintf("%s/cars/%s", apiBaseURL, id))
	if err != nil {
		return Car{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Car{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var car Car
	if err := json.NewDecoder(resp.Body).Decode(&car); err != nil {
		return Car{}, err
	}

	return car, nil
}

// createCar creates a new car via the API
func createCar(car Car) error {
	payload, err := json.Marshal(car)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/cars", apiBaseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateCar updates an existing car via the API
func updateCar(car Car) error {
	payload, err := json.Marshal(car)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/cars/%s", apiBaseURL, car.ID),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// deleteCar deletes a car via the API
func deleteCar(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/cars/%s", apiBaseURL, id), nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
