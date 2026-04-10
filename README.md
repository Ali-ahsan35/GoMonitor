# GoMonitor

A full-stack API performance and profiling dashboard built with Go and React.

Repository: https://github.com/Ali-ahsan35/GoMonitor

Live Frontend: https://gomonitor.netlify.app

## What This Project Does

GoMonitor helps you test multiple APIs at once and understand both:

- External API behavior: latency, status codes, success/failure
- Internal Go runtime behavior: CPU, memory, goroutines, mutex/block contention

In short, it answers:

- Which API endpoints are slow or failing?
- Why is my backend process consuming CPU/memory?

## Key Features

### API Latency Dashboard

- Test multiple URLs concurrently using goroutines
- Optional custom request headers (for protected endpoints)
- Per-URL metrics:
	- Response time (ms)
	- HTTP status code
	- Success/failure
- Run summary:
	- Total execution time
	- Success/failure counts
	- Active goroutine count
	- Memory allocation

### Profiling Dashboard

- Capture profile summaries from backend
- Supported profile types:
	- `cpu`
	- `heap`
	- `allocs`
	- `goroutine`
	- `mutex`
	- `block`
	- `threadcreate`
- Visualize top hotspots in charts/tables
- Download raw `.pprof` files for deeper analysis with `go tool pprof`

## Tech Stack

### Backend

- Go
- Gin
- Goroutines + channels + sync.WaitGroup
- pprof (`gin-contrib/pprof`)

### Frontend

- React (Vite)
- Axios
- Recharts

## Project Structure

```text
GoMonitor/
|- Backend/
|  |- cmd/server/
|  |- handler/
|  |- model/
|  |- pkg/
|  |  |- httpclient/
|  |  |- metrics/
|  |  |- profiler/
|  |- service/
|- Frontend/
|  |- src/
|- README.md
```

## Quick Start

### 1. Clone

```bash
git clone https://github.com/Ali-ahsan35/GoMonitor.git
cd GoMonitor
```

### 2. Run Backend

```bash
cd Backend
go mod tidy
go run ./cmd/server
```

Backend runs at `http://localhost:8080`.

### 3. Run Frontend

Open a new terminal:

```bash
cd Frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`.

Optional custom backend URL:

```bash
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```

## API Endpoints

### Health

- `GET /health`

### API Testing

- `POST /run-test`

Request body:

```json
{
	"urls": [
		"https://api-1.example.com",
		"https://api-2.example.com"
	],
	"headers": {
		"x-api-key": "YOUR_API_KEY",
		"X-Requested-With": "XMLHttpRequest"
	}
}
```

Response shape:

```json
{
	"results": [
		{
			"url": "https://api-1.example.com",
			"time": 120,
			"status": 200,
			"success": true
		}
	],
	"summary": {
		"total_time": 500,
		"success_count": 1,
		"failure_count": 1,
		"goroutines": 5,
		"memory_alloc": 123456
	}
}
```

### Profiling

- `GET /debug/pprof/` and standard pprof subroutes
- `GET /profiles/types`
- `POST /profiles/capture`
- `GET /profiles/:type/download`

Capture request example:

```json
{
	"type": "cpu",
	"seconds": 10
}
```

Notes:

- `seconds` is mainly relevant for `cpu` capture.
- For download: use `GET /profiles/cpu/download?seconds=10` for timed CPU profile capture.

## Example Use Cases

1. Compare multiple API variants (regions, query params, filters).
2. Validate protected endpoint access with custom headers.
3. Detect slow endpoints before release.
4. Profile backend hotspots when scaling traffic.

## Running pprof Locally

After downloading a profile file:

```bash
go tool pprof /path/to/profile.pprof
```

Useful commands inside pprof:

- `top`
- `list <function-name>`
- `web` (if graphviz is installed)

## Troubleshooting

### Port 8080 already in use

If backend fails with `bind: address already in use`:

```bash
lsof -i :8080 -n -P
kill <PID>
```

Then rerun backend.

### 401 responses from tested APIs

Provide required headers in the dashboard form or request payload, for example:

- `x-api-key`
- `X-Requested-With`

### Empty/low profiling data

Generate some traffic first (run test requests), then capture profile again.

## Roadmap

- Persistent test/profile history
- Compare multiple runs over time
- Export summary as CSV/JSON
- WebSocket live updates during test runs


