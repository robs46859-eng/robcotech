# FullStackArkham

A modular AI operating system built around a centralized inference gateway, persistent orchestration, low-cost model routing, semantic caching, evolving memory, and reusable vertical applications.

## Architecture

This is a workspace-based monorepo following the architecture defined in the [Master Architecture Document](../agent_init/master_architecture_doc.md).

### Core Components

1. **Gateway and Control Plane** (`services/gateway`) - Central policy and inference control plane
2. **BIM Ingestion Service** (`services/bim_ingestion`) - IFC parsing and normalization
3. **Orchestration Layer** (`services/orchestration`) - Event-driven workflow execution
4. **Memory Service** (`services/memory`) - A-MEM evolving memory system
5. **Semantic Cache** (`services/semantic-cache`) - Request/response caching with embeddings
6. **Billing** (`services/billing`) - Usage metering and Stripe integration

### Vertical Applications

- **Web App** (`apps/web`) - Next.js frontend for BIM workflows
- **Admin** (`apps/admin`) - Tenant management, usage monitoring, alerts
- **Docs** (`apps/docs`) - Documentation site

## Phase 1: Core Foundation

Current focus:
- [x] Repository structure
- [ ] Gateway skeleton with Go
- [ ] Tenant auth and rate limiting
- [ ] Health and readiness endpoints
- [ ] Postgres, Redis, vector store setup

## Tech Stack

- **Gateway**: Go (following adk-go patterns)
- **Frontend**: Next.js 16, React, TypeScript
- **Database**: PostgreSQL with PostGIS
- **Cache**: Redis
- **Vector Store**: LanceDB/FAISS
- **Queue**: Redis-based (BullMQ compatible)

## License

AGPL v3 - See [LICENSE](LICENSE)
