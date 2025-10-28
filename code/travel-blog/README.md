# Travel Blog Platform

This project contains a small travel CMS and public site for recording visited countries and places.

## Components
- **Backend**: Go (Gin) API that manages countries and places in a Postgres database.
- **Frontend**: Static HTML/CSS/JS experience served through Nginx. Includes a public gallery and lightweight admin forms.
- **Infrastructure**: Docker Compose orchestration with Postgres and containerized services.

### Frontend layout
- **Public UI**: The "Destinations" section in `frontend/index.html` renders the public catalogue of countries and the places within them. It is driven by the country and place templates in the same file and the rendering helpers in `frontend/main.js`.
- **Admin UI**: The "Content Management" section in `frontend/index.html` exposes the country and place submission forms. Form handling logic lives alongside the fetch helpers in `frontend/main.js`.

## Running locally
```bash
docker compose up --build
```

The frontend is available at <http://localhost:8088>. API endpoints are proxied through `/api`.

Set `DATABASE_URL` when running the backend outside Docker, for example:
```bash
export DATABASE_URL="postgres://travel:travel@localhost:5432/travel?sslmode=disable"
go run ./code/travel-blog/backend/cmd/server
```
