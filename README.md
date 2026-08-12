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
