# FullStackArkham - How to Run E2E Tests

## Quick Start

**Prerequisites:**
- Docker Desktop running
- Go 1.25+ installed
- Python 3.11+ installed

---

## Step 1: Build Go Gateway for Linux

```bash
cd /Users/joeiton/Desktop/FullStackArkham/services/gateway
GOOS=linux GOARCH=amd64 go build -o bin/gateway ./app
```

This creates a Linux-compatible binary for the Docker container.

---

## Step 2: Start All Services

```bash
cd /Users/joeiton/Desktop/FullStackArkham
docker-compose up -d
```

Wait 15 seconds for services to start.

---

## Step 3: Check Service Status

```bash
docker-compose ps
```

You should see:
- fullstackarkham-postgres (healthy)
- fullstackarkham-redis (healthy)
- fullstackarkham-gateway (healthy)
- fullstackarkham-arkham (healthy)

---

## Step 4: Run E2E Tests

```bash
python3 tests/run_e2e_test.py
```

---

## Expected Test Results

### Database Test
```
✓ Database connected successfully
  - Host: localhost:15432
  - Database: fullstackarkham
  - Tenants in database: 0
```

### Health Checks
```
✓ gateway: healthy
✓ arkham: healthy
✓ bim_ingestion: healthy (if built)
✓ orchestration: healthy (if built)
...
```

### Gateway Inference
```
✓ Gateway inference successful
  - Model: local/phi-2
  - Response preview: I'm a local model...
```

### Arkham Security
```
✓ Arkham classification successful
  - Classification: benign
  - Threat score: 0.10
  - Recommended action: pass
```

---

## Troubleshooting

### Gateway won't start
```bash
# Check logs
docker logs fullstackarkham-gateway

# Rebuild for Linux
cd services/gateway
GOOS=linux GOARCH=amd64 go build -o bin/gateway ./app
docker-compose restart gateway
```

### Port conflicts
If ports 15432, 16379, 8080, etc. are in use, edit `docker-compose.yml` to use different ports.

### Python services fail
```bash
# Check logs
docker logs fullstackarkham-arkham

# Rebuild with dependencies
docker-compose up -d --build arkham
```

### Database connection fails
```bash
# Check postgres is running
docker-compose ps postgres

# Check logs
docker logs fullstackarkham-postgres
```

---

## Service Ports

| Service | Host Port | Container Port |
|---------|-----------|----------------|
| Gateway | 8080 | 8080 |
| Arkham | 8081 | 8080 |
| BIM Ingestion | 8082 | 8080 |
| Orchestration | 8083 | 8080 |
| Semantic Cache | 8084 | 8080 |
| Memory | 8085 | 8080 |
| Billing | 8086 | 8080 |
| PostgreSQL | 15432 | 5432 |
| Redis | 16379 | 6379 |

---

## Clean Up

```bash
# Stop all services
docker-compose down

# Remove volumes (reset database)
docker-compose down -v
```

---

*Created: April 27, 2026*
