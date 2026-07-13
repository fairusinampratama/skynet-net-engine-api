# Skynet NetEngine API 🚀

**NetEngine** is a high-performance middleware designed for ISP management. It acts as the "Muscle" between your "Brain" (e.g., Laravel/PHP App) and your MikroTik router fleet.

It maintains persistent, self-healing TCP connections to hundreds of routers, allowing for millisecond-latency commands and real-time monitoring without the overhead of establishing new connections for every request.

## ✨ Features

### Backend
*   **Persistent Connections**: Dedicated Goroutine per router with automatic reconnection logic
*   **High Concurrency**: Thread-safe worker pool with serialized router operations
*   **Startup Warmup**: Ensures data is cached before serving traffic (eliminates race conditions)
*   **Smart Queue Detection**: Automatic fallback for PPPoE queue naming patterns
*   **REST API Bridge**: Simple HTTP endpoints for complex RouterOS commands
*   **Real-time Monitoring**: CPU, Memory, and live Queue Traffic stats with 10s timeout
*   **Enterprise Security**: API Key authentication (`X-App-Key`) and Webhook event dispatching
*   **Swagger Documentation**: Built-in interactive API docs at `/api/v1/swagger/index.html`

### Dashboard (Frontend)
> **Note**: The React frontend has been moved to the `feature/full-stack` branch. The `main` branch is now a dedicated Go backend API.

To run the full stack (Backend + Frontend), checkout the branch:
```bash
git checkout feature/full-stack
```

## 🛠️ Tech Stack

*   **Language**: Go (Golang) 1.22+
*   **Framework**: Gin Gonic
*   **Database**: MySQL (for router credentials)
*   **MikroTik Lib**: `go-routeros`
*   **Logging**: Uber Zap

## 🚀 Deployment

### Option 1: Native Go Build (Recommended)
This is the simplest way to run the backend engine.

```bash
# 1. Build
go build -o skynet-net-engine-api cmd/server/main.go

# 2. Run
./skynet-net-engine-api
```

### Option 2: Nixpacks / Coolify
This project includes a `nixpacks.toml` for zero-config deployment.
If you are using **Coolify** or any Nixpacks-compatible platform, simply push this repository and it will build automatically.

### Option 3: Systemd (Linux Service)
To run in background as a service:

1. Create `/etc/systemd/system/net-engine.service`:
```ini
[Unit]
Description=Skynet Net Engine
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/app
ExecStart=/path/to/app/skynet-net-engine-api
Restart=always
EnvironmentFile=/path/to/app/.env

[Install]
WantedBy=multi-user.target
```

2. Enable and Start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable net-engine --now
```

## ⚙️ Configuration

Create a `.env` file (or rely on defaults/docker-compose):

```ini
DB_DSN="username:password@tcp(127.0.0.1:3306)/netengine?parseTime=true"
API_PORT=":8080"
APP_KEY="<generate-a-long-random-secret>"
```

Do not commit real router credentials, database credentials, API keys, `.env` files, exported router data, or production network details. If any real credential is committed to a public branch, rotate it immediately and purge it from Git history.

## 🖥️ Dashboard

Access the web dashboard at: **[http://localhost:8080](http://localhost:8080)**

Features:
- Real-time traffic graphs with click-to-monitor
- System resource monitoring (CPU/Memory/Uptime)
- Active user sessions with pagination
- Multi-router selector dropdown
- Automatic skeleton loaders during data fetch

## 📚 API Documentation

Once running, access the full Swagger UI at:
**[http://localhost:8080/api/v1/swagger/index.html](http://localhost:8080/api/v1/swagger/index.html)**

### Core Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/` | Web Dashboard (React App) |
| `GET` | `/api/v1/health` | System health check |
| `GET` | `/api/v1/routers` | List all configured routers |
| `GET` | `/api/v1/monitoring/targets` | Get all active users |
| `GET` | `/api/v1/router/:id/health` | Router CPU/Memory stats |
| `GET` | `/api/v1/router/:id/traffic?user=USERNAME` | Live user traffic (bits/sec) |
| `POST` | `/api/v1/sync/:id` | Force router sync |
| `POST` | `/api/v1/secret` | Create PPPoE secret |
| `POST` | `/api/v1/isolate` | Isolate/Unisolate customer |

**Authentication**: All `/api/v1/*` routes require the `X-App-Key` header. Use a long random value from your private environment, never a value committed to the repository.

## 🧪 Development

### Running Tests
```bash
go test ./... -v
```

### Database Seeding
To populate the database with initial router data:
```bash
go run cmd/seeder/main.go
```

## 🏗️ Architecture

### Backend Flow
```
HTTP Request → Gin Router → API Handler → Worker Pool → MikroTik Router
                                              ↓
                                         Command Queue (Buffered Channel)
                                              ↓
                                         Worker Goroutine (Serialized)
                                              ↓
                                         RouterOS API Response
```

### Key Improvements
- **Warmup Phase**: Server blocks until routers connect and cache initial data
- **Thread Safety**: All router operations serialized through command queue
- **PPPoE Detection**: Smart fallback for `<pppoe-USERNAME>` queue naming
- **Error Resilience**: Frontend treats 500-504 errors as transient loading states

## 📝 License
Proprietary / Internal Use Only.

## Public Portfolio Note

This repository documents the backend architecture and implementation approach. Public materials should avoid exposing production router details, customer identifiers, live credentials, internal IP addresses, or operational diagrams.
