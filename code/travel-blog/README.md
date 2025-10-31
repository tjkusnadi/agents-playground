# Travel Blog Platform

This project contains a small travel CMS and public site for recording visited countries and places.

## Components
- **Backend**: Go (Gin) API that manages countries and places in a Postgres database.
- **Public Frontend**: Static HTML/CSS/JS experience served through Nginx for browsing destinations.
- **Admin Frontend**: Separate Nginx site that surfaces the content management forms.
- **Infrastructure**: Docker Compose orchestration with Postgres and containerized services.

### Frontend layout
- **Public UI**: Lives in `frontend/`. The "Destinations" section in `frontend/index.html` renders the public catalogue of countries and the places within them, driven by the templates in the same file and the rendering helpers in `frontend/main.js`.
- **Admin UI**: Lives in `frontend-admin/`. The admin dashboard mirrors the destination list while exposing country and place submission forms. Form handling logic and API helpers live in `frontend-admin/main.js`.

## Running locally
```bash
docker compose -f code/travel-blog/docker-compose.yml up --build
```

The public frontend is available at <http://localhost:8088> and the admin dashboard at <http://localhost:8090>. API endpoints are proxied through `/api` in both frontends.

Set `DATABASE_URL` when running the backend outside Docker, for example:
```bash
export DATABASE_URL="postgres://travel:travel@localhost:5432/travel?sslmode=disable"
go run ./code/travel-blog/backend/cmd/server
```
