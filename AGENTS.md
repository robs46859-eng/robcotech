# FullStackArkham - AI Assistant Guide

## Project Overview

FullStackArkham is a modular AI operating system built around:
- Centralized inference gateway (Go)
- Persistent orchestration
- Low-cost model routing
- Semantic caching
- Evolving memory (A-MEM)
- BIM vertical application

## Repository Structure

```
FullStackArkham/
├── apps/                    # Frontend applications
│   ├── web/                 # Next.js BIM frontend
│   ├── admin/               # Admin dashboard
│   └── docs/                # Documentation
├── services/                # Backend microservices
│   ├── gateway/             # Go inference gateway (Layer8)
│   ├── bim_ingestion/       # Python IFC parsing service
│   ├── orchestration/       # Workflow execution
│   ├── memory/              # A-MEM memory service
│   ├── semantic-cache/      # Semantic caching
│   ├── billing/             # Usage metering
│   └── arkham/              # Active defense & deception
├── packages/                # Shared packages
│   ├── shared-types/        # Common type definitions
│   └── schemas/             # JSON schemas
├── infra/                   # Infrastructure
│   ├── docker/              # Docker configs
│   ├── k8s/                 # Kubernetes manifests
│   └── terraform/           # Terraform configs
└── docker-compose.yml       # Local development
```

## Key Architecture Decisions

### Gateway Language: Go
- Following adk-go patterns from `agent_init/adk-go`
- Better performance for high-throughput inference routing
- Strong typing and concurrency model

### BIM-First Vertical
- First vertical proof: IFC parsing → project status artifact
- Uses PostGIS for spatial data
- Heavy parsing queued asynchronously

### Arkham Security Integration
- Active defense with deception layer
- Cross-tenant attacker fingerprint sharing
- Behavioral fingerprinting (not just signatures)

### Cost Ladder
1. Deterministic systems first (rules, schema validation)
2. Local models second (routing, classification)
3. Low-cost API models third (moderate reasoning)
4. Premium API models last (complex synthesis)

## Development Workflow

### Local Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f gateway

# Run migrations
docker-compose exec postgres psql -U postgres -d fullstackarkham
```

### Key Endpoints

| Service | Port | Endpoint |
|---------|------|----------|
| Gateway | 8080 | POST /v1/ai |
| BIM Ingestion | 8082 | POST /api/v1/projects |
| Orchestration | 8083 | POST /api/v1/workflows |
| Memory | 8085 | POST /api/v1/notes |

## Database Schema

Core tables in `infra/docker/postgres/init.sql`:
- `tenants` - Multi-tenant isolation
- `users` - User accounts
- `api_keys` - API key management
- `inference_logs` - All inference requests
- `workflows` - Orchestration state
- `semantic_cache` - Cached responses
- `memory_notes` - A-MEM notes
- `arkham_events` - Security events

## AI Assistant Guidelines

### When Building New Features

1. **Check existing patterns** in `agent_init/` repos:
   - Gateway patterns: `agent_init/adk-go`
   - Security patterns: `agent_init/arkham`
   - BIM patterns: `agent_init/fullstack-bim`
   - Layer8 patterns: `agent_init/layer8`

2. **Follow the cost ladder** - don't use premium models for simple tasks

3. **Persist workflow state** - all multi-step operations must be resumable

4. **Tenant-aware** - every query must include tenant_id

5. **Arkham integration** - all requests pass through security layer

### Graphify Integration

This project uses graphify for knowledge management:
- Run `graphify update .` after code changes
- Read `graphify-out/GRAPH_REPORT.md` for architecture questions
- Use `graphify query "<question>"` for specific queries

## Phase 1 Priorities

**Knowledge Graph:** 846 nodes, 1,578 edges, 32 communities

### Completed Services
- [x] Repository structure (monorepo per spec section 17)
- [x] Gateway skeleton (Go) with model routing (cost ladder)
- [x] Gateway ↔ Arkham integration (threat classification, deception, blocks)
- [x] Gateway ↔ Provider integration (Anthropic, OpenAI, Google, Local)
- [x] Production auth (PostgreSQL tenant store with bcrypt API keys)
- [x] Database schema (14 tables + billing/subscription)
- [x] BIM ingestion service (IFC parsing with IfcOpenShell)
- [x] Arkham security service (detector, fingerprint, deception, audit)
- [x] Orchestration service (flow registry, checkpoints, task queues, executors)
- [x] Memory service (A-MEM notes, retrieval, links, evolution)
- [x] Semantic cache service (embeddings, lookup, storage)
- [x] Billing service (metering, plans, Stripe integration)
- [x] Next.js web frontend (project upload, status dashboard)
- [x] Docker Compose (8 services + Postgres + PostGIS + Redis)
- [x] Kubernetes configs (gateway, arkham, postgres deployments)
- [x] Terraform configs (GCP: GKE, Cloud SQL, Memorystore)
- [x] Test suite (e2e tests, unit tests for routing/detector)
- [x] Makefile for common operations

### Production Ready
All core services are implemented. To deploy:

```bash
# Local development
make dev

# Kubernetes deployment
kubectl apply -f infra/k8s/

# GCP infrastructure
cd infra/terraform && terraform apply
```

## Non-Negotiable Rules

From Master Architecture Document Section 14:

1. No direct model calls from vertical apps
2. No duplicated auth, memory, or caching logic
3. No premium model by default
4. No side effects without idempotency
5. No unscoped memory
6. No cross-tenant cache leakage
7. No workflow without persisted checkpoints
8. No production without logs, metrics, failure visibility
9. No schema-free outputs at critical boundaries
10. No multiple verticals before first one proves backbone
