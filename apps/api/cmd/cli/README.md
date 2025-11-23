# CarFlow CLI

A command-line interface for interacting with the CarFlow API.

## Building the CLI

You can build the CLI tool with:

```bash
make build-cli
```

This will create a `carflow-cli` executable in the project root.

## Usage

```
CarFlow CLI - A command-line interface for the CarFlow API

Usage:
  carflow-cli [command] [options]

Commands:
  list    - List all cars with optional filtering and pagination
  get     - Get a specific car by ID
  create  - Create a new car
  update  - Update an existing car
  delete  - Delete a car
  health  - Check API health
  help    - Show this help message

Run 'carflow-cli [command] -h' for more information on a command.
```

## Examples

### Listing cars

List all cars:
```bash
./carflow-cli list
```

List cars with pagination:
```bash
./carflow-cli list -page 1 -page-size 2
```

Filter cars by make:
```bash
./carflow-cli list -make "Toyota"
```

Filter and sort cars:
```bash
./carflow-cli list -make "Toyota" -sort "year" -order "desc"
```

### Getting a specific car

```bash
./carflow-cli get -id "1"
```

### Creating a car

```bash
./carflow-cli create -id "new-car" -make "BMW" -model "X5" -year 2023 -color "black"
```

### Updating a car

```bash
./carflow-cli update -id "1" -color "silver"
```

### Deleting a car

```bash
./carflow-cli delete -id "1"
```

### Checking API health

```bash
./carflow-cli health
``` 