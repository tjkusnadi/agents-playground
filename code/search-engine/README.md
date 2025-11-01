# Search Engine Agent

This project implements the Search Engine agent requirements with a Go (Gin) backend, Elasticsearch for data storage and search, and a lightweight vanilla JavaScript frontend for interacting with the API.

## Prerequisites

- Go 1.22+
- Node.js is **not** required; the frontend is plain HTML/CSS/JS served by Gin.
- Elasticsearch 8.x instance accessible to the backend (for local development a Docker container is easiest).

### Starting Elasticsearch with Docker

```bash
docker run --rm \
  --name elasticsearch \
  -p 9200:9200 \
  -e discovery.type=single-node \
  -e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
  docker.elastic.co/elasticsearch/elasticsearch:8.12.2
```

The backend defaults to `http://localhost:9200`. Set the following variables if you need to override them:

- `ELASTICSEARCH_ADDRESS`
- `ELASTICSEARCH_USERNAME`
- `ELASTICSEARCH_PASSWORD`

## Running the backend + frontend

```bash
cd code/search-engine/backend
go run .
```

The server starts on `http://localhost:8080`. On startup it will:

1. Ensure the `movies` index exists with the correct mapping.
2. Seed the index with five sample movies if it is empty.
3. Serve the API under `/api` and the static frontend at `/` (served from `../frontend`).

If you prefer to host the frontend separately, set `FRONTEND_DIR` to the location of the static files or serve them via another server and point API calls to the backend URL.

## Running everything with Docker

The repository includes a Dockerfile for the Go backend (with the static frontend assets baked in), an nginx reverse proxy, and an Elasticsearch container orchestrated through Docker Compose.

### Prerequisites

- Docker Engine 24+
- Docker Compose v2 (`docker compose` CLI)

### Start the stack

From the `code/search-engine` directory run:

```bash
docker compose up --build
```

The command builds the backend image, starts Elasticsearch, and then launches nginx. Once healthy, visit `http://localhost:8080` to use the UI. API requests (`/api/...`) are proxied by nginx to the Go service running inside the `backend` container. The Elasticsearch API remains available on `http://localhost:9200` if you need to inspect the index directly.

### Tear everything down

```bash
docker compose down
```

Add `-v` to also remove the ephemeral Elasticsearch data volume.

## API Overview

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| `GET` | `/api/movies` | Search movies with optional `q`, `page`, and `pageSize` parameters. |
| `GET` | `/api/movies/:id` | Retrieve a single movie document. |
| `POST` | `/api/movies` | Create a new movie. |
| `PUT` | `/api/movies/:id` | Replace a movie document (supply all fields). |
| `DELETE` | `/api/movies/:id` | Delete a movie by id. |

All write operations immediately refresh the index to make documents available to search.

## Frontend Features

- Search bar with adjustable page size and server-side pagination controls.
- Result list showing title, genre, rating, release year, description, and document ID.
- Management forms to create, update (with a load button that fetches the latest data), and delete movies.

The frontend communicates with the backend via `fetch` using relative paths, so it will work as long as the API is accessible under the same origin or proxied accordingly.
