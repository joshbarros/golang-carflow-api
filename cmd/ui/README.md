# CarFlow UI

A simple web interface for the CarFlow API, built using Go's standard library templates.

## Features

- View all cars with filtering, sorting, and pagination
- View details of a specific car
- Create new cars
- Edit existing cars
- Delete cars
- Check API health status

## Building and Running

You can build and run the UI with:

```bash
# Build the UI application
make build-ui

# Run the UI application
make run-ui
```

By default, the UI will be available at http://localhost:3000 and will connect to the CarFlow API at http://localhost:8080.

## Structure

The UI application is organized as follows:

- `cmd/ui/main.go`: The main application file
- `cmd/ui/templates/`: HTML templates
  - `layout.html`: Base layout template 
  - `home.html`: Home page
  - `list.html`: Car listing page with filtering and pagination
  - `view.html`: Car details page
  - `new.html`: Create car form
  - `edit.html`: Edit car form
  - `delete.html`: Delete car confirmation
  - `error.html`: Error display

## Development

To modify the UI:

1. Edit the Go code in `cmd/ui/main.go` for functionality changes
2. Edit the HTML templates in `cmd/ui/templates/` for appearance changes
3. Run with `make run-ui` to see your changes

## Screenshots

Screenshots will be added here once the UI is fully implemented.

## Dependencies

The UI uses:
- Go standard library
- Bootstrap 5 (loaded from CDN) 