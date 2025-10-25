# Currency Converter Agent

This project contains a minimal full-stack implementation for the currency-converter agent task.

## Backend (Go)

* Location: `backend/`
* Framework: Go standard library (`net/http`)
* Endpoints:
  * `GET /api/convert?base=<BASE>&target=<TARGET>&amount=<AMOUNT>` — proxies conversion rates from Yahoo Finance and returns the converted amount.
  * `GET /healthz` — simple health-check endpoint.
* Environment: listens on port `8080` by default (can be overridden with the `PORT` environment variable).

## Frontend (React + TypeScript)

* Location: `frontend/`
* Tooling: [Vite](https://vitejs.dev/) for development and build.
* Scripts:
  * `npm run dev` — start the development server with proxying to the Go backend.
  * `npm run build` — type-check and build for production.
  * `npm run preview` — preview the production build.

## Development

1. Start the Go backend:

   ```bash
   cd backend
   go run .
   ```

2. In a separate terminal, start the React dev server:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

The Vite dev server proxies API calls under `/api` to `http://localhost:8080` so the frontend can communicate with the backend seamlessly.
