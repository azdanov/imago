# Imago

## Overview

This is a demo application built with Go to explore fullstack development. It uses html templates and postgres as a database.

## Getting Started

### Prerequisites

- Go (1.24) installed: [https://golang.org/dl/](https://golang.org/dl/)
- Docker installed: [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)
- Make installed: [https://www.gnu.org/software/make/](https://www.gnu.org/software/make/)

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/azdanov/imago.git
    cd imago
    ```
2.  Install dependencies:
    ```bash
    go mod tidy
    ```
3. Initialize the project:
    ```bash
    make init
    ```

### Running the Application

```bash
go run main.go
# Or using make (to start with air for hot reload)
make dev
# To see all make commands
make help
```

## Usage

This application provides a simple web interface for managing images and galleries. You can upload, view, and delete images. To use it you need to create a user.

## License

This project is licensed under the MIT License.

