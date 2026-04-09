# GoMonitor

GoMonitor is a full-stack API performance dashboard.

## Structure

- `Backend/`: Go + Gin API with concurrency testing and pprof
- `Frontend/`: React (Vite) dashboard with charts and summary cards

## Backend

```bash
cd Backend
go mod tidy
go run ./cmd/server
```

Backend runs at `http://localhost:8080`.

### Endpoints

- `POST /run-test`
- `GET /health`
- `GET /debug/pprof/` (and other pprof routes)
- `GET /profiles/types`
- `POST /profiles/capture`
- `GET /profiles/:type/download`

Example `POST /run-test` body with optional headers:

```json
{
	"urls": [
		"https://ownerdirect.beta.123presto.com/api/v1/category/details/usa:hawaii?amenities=19&device=desktop&items=1&limit=8&locations=US&showFallbackData=1"
	],
	"headers": {
		"x-api-key": "YOUR_API_KEY",
		"X-Requested-With": "XMLHttpRequest"
	}
}
```

Example `POST /profiles/capture` body:

```json
{
	"type": "cpu",
	"seconds": 10
}
```

Supported profile types: `cpu`, `heap`, `allocs`, `goroutine`, `mutex`, `block`, `threadcreate`

## Frontend

```bash
cd Frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`.

The frontend includes two modes:

- Latency dashboard for API response metrics
- Profiling dashboard for CPU/heap/goroutine/mutex/block capture summaries and hotspot charts

To customize API base URL:

```bash
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```
