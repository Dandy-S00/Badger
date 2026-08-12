# Badger API Gateway

A fully deployment-ready API discovery gateway available in both **Go** and **Python**.

## Features

- **GET /health** — Liveness/readiness probe for container orchestration
- **GET /apis** — Lists all registered routes with metadata (auth, rate limit, tags)
- **GET /openapi.json** — Dynamically generates an OpenAPI 3.0 specification
- Auth-protected CRUD endpoints with Bearer token middleware
- Postman collection with smoke-test scripts
- Docker + Docker Compose — single-command deployment
- GitHub Actions CI/CD — test, build, publish, and verify

- ## Endpoints Summary

| Method | Path | Auth Required | Purpose |
| --- | --- | --- | --- |
| GET | /health | No | Health check for probes |
| GET | /apis | No | List all discovered routes |
| GET | /openapi.json | No | OpenAPI 3.0 specification |
| GET | /api/v1/data | Yes (Bearer) | Protected example data |
| GET | /api/v1/users | Yes (Bearer) | List users |
| POST | /api/v1/users | Yes (Bearer) | Create a user |

# .github/workflows/ci.yml
- Test Go: go vet + build
- Test Python: py_compile
- Build & push Docker images (GitHub Container Registry)
- Smoke test: curl /apis, /openapi.json
- Run Postman collection via Newman action
