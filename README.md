```
  ██████╗ ██╗████████╗    ██████╗  █████╗ ███╗   ██╗███╗   ██╗███████╗██████╗
 ██╔════╝ ██║╚══██╔══╝    ██╔══██╗██╔══██╗████╗  ██║████╗  ██║██╔════╝██╔══██╗
 ██║  ███╗██║   ██║       ██████╔╝███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
 ██║   ██║██║   ██║       ██╔══██╗██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
 ╚██████╔╝██║   ██║       ██████╔╝██║  ██║██║ ╚████║██║ ╚████║███████╗██║  ██║
  ╚═════╝ ╚═╝   ╚═╝       ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
```

# git-banner-backend

A Go backend that generates SVG banners from live GitHub profile data, sized for Twitter and LinkedIn.

## Features

- Three banner types — **stats**, **contributions heatmap**, **pinned repositories**
- Two platform formats — **Twitter** (1500×500) and **LinkedIn** (1584×396)
- Two themes — **dark** and **light**
- Self-contained SVGs — avatars are embedded as base64 data URIs (no broken images when downloaded)
- In-memory response cache with configurable TTL
- Per-IP rate limiting
- Zero external dependencies — pure Go standard library

---

## API

### `GET /banner/{username}`

Generates an SVG banner for the given GitHub username.

| Parameter | Type | Default | Values |
|---|---|---|---|
| `type` | query | `stats` | `stats` \| `contributions` \| `pinned` |
| `format` | query | `twitter` | `twitter` \| `linkedin` |
| `theme` | query | `dark` | `dark` \| `light` |

**Response:** `image/svg+xml` with `Cache-Control: public, max-age=300`

> `contributions` and `pinned` types require `GITHUB_TOKEN` to be set (GitHub GraphQL API restriction).

---

### `GET /health`

Returns `{"status":"ok"}` with HTTP 200 when the server is running.

---

## Banner Types

| Type | What it shows |
|---|---|
| `stats` | Followers, following, public repos, total stars |
| `contributions` | Full-year contribution heatmap with month/day labels |
| `pinned` | Up to 6 pinned repositories with description, language, stars, forks |

---

## Running Locally

### Prerequisites
- Go 1.23+
- A GitHub personal access token (optional, but strongly recommended)

### Setup

```bash
git clone https://github.com/somya/git-banner-backend.git
cd git-banner-backend

cp .env.example .env
# Edit .env and add your GITHUB_TOKEN
```

### Start the server

```bash
make dev
# or
make run
```

Server starts on `http://localhost:8080`.

### Available make targets

```
make build   — compile binary to bin/
make run     — build and run
make dev     — run with go run (no binary)
make test    — run all tests
make lint    — run go vet
make fmt     — format source files
make tidy    — tidy and verify go modules
make clean   — remove compiled binaries
```

---

## Configuration

All configuration is via environment variables (or a `.env` file in the working directory).

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port the server listens on |
| `ENV` | `development` | `development` or `production` |
| `GITHUB_TOKEN` | _(none)_ | GitHub PAT — 5,000 req/hr vs 60 without |
| `CACHE_TTL` | `300` | Seconds to cache GitHub API responses |
| `RATE_LIMIT` | `60` | Max requests per minute per IP (0 = unlimited) |

---

## Deployment

### Docker

```bash
docker build -t git-banner-backend .
docker run -p 8080:8080 -e GITHUB_TOKEN=ghp_... git-banner-backend
```
