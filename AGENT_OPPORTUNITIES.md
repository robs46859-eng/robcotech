# Agent Build Opportunities

This document identifies specific areas where an AI coding agent (like the one building this system) could improve the FullStackArkham codebase. These are intentional gaps left for agent-driven enhancement.

---

## High Priority - Core Functionality

### 1. Gateway Model Routing Implementation
**Location:** `services/gateway/app/api/v1/handlers.go`

**Current State:** Inference handler returns placeholder response

**Agent Task:** Implement the full inference pipeline:
- Integrate with semantic cache service (check before inference)
- Implement model routing policy (cost ladder: cheap → premium)
- Add provider integration (Anthropic, OpenAI, Google via adk-go patterns)
- Wire up fallback chains when primary provider fails

**Reference:** `agent_init/adk-go` for provider patterns, `agent_init/layer8` for routing logic

**Priority:** ⭐⭐⭐ Critical for Phase 2

---

### 2. Arkham Detector Implementation
**Location:** `services/arkham/app/detector.py` (not created)

**Current State:** ThreatDetector class referenced but not implemented

**Agent Task:** Build threat classification system:
- Implement MITRE ATT&CK technique detection
- Behavioral pattern analysis (request frequency, header anomalies, payload structure)
- Confidence scoring for classifications
- Integration with gateway middleware

**Reference:** `agent_init/arkham/services/arkham/detector/` for existing patterns

**Priority:** ⭐⭐⭐ Critical for security layer

---

### 3. Arkham Deception Generator
**Location:** `services/arkham/app/deception.py` (not created)

**Current State:** DeceptionGenerator class referenced but not implemented

**Agent Task:** Build AI deception system:
- Generate novel trap configurations per session
- Create plausible false API responses (fake auth success, rate limit messages)
- Implement trap types: honeypot endpoints, fake schemas, dead-end sessions
- Ensure no two traps are identical for same fingerprint

**Reference:** `agent_init/arkham/arkham-security-strategy.docx.md` section 2

**Priority:** ⭐⭐⭐ Core innovation

---

### 4. BIM IFC Parser Production Implementation
**Location:** `services/bim_ingestion/app/parsers/ifc.py`

**Current State:** Has mock parser fallback, real IfcOpenShell integration incomplete

**Agent Task:** Complete IFC parsing:
- Full entity extraction (walls, doors, windows, spatial structure)
- Property set and quantity extraction
- Material associations
- Issue detection (unnamed elements, duplicate GUIDs, missing data)
- Test with `agent_init/BIMS_Structural.ifc`

**Priority:** ⭐⭐⭐ BIM-first vertical depends on this

---

### 5. Workflow Task Executors
**Location:** `services/orchestration/app/tasks/` (directory not created)

**Current State:** Flow registry defines task types but no executors

**Agent Task:** Implement task executors for each task type:
- `bim_retrieval` - Query BIM domain store
- `memory_retrieval` - Call memory service
- `model_inference` - Call gateway with model tier
- `policy_evaluation` - Apply routing policies
- `schema_validation` - Validate outputs against JSON schemas
- `artifact_storage` - Save to object storage

**Priority:** ⭐⭐⭐ Orchestration can't execute without these

---

## Medium Priority - Integration

### 6. Gateway ↔ Arkham Integration
**Location:** `services/gateway/app/middleware/middleware.go`

**Current State:** `ArkhamMiddleware` is a placeholder

**Agent Task:** Wire up Arkham security:
- HTTP client calls to Arkham classify endpoint
- Handle benign → pass, probe → deceive, attack → block
- Add deception response handling
- Log all security events to audit table

**Priority:** ⭐⭐ Security integration

---

### 7. Semantic Cache Embedding Service
**Location:** `services/semantic-cache/app/embeddings.py`

**Current State:** Uses mock embeddings, sentence-transformers optional

**Agent Task:** Production embedding setup:
- Install and configure sentence-transformers or use API embeddings
- Implement proper cosine similarity search
- Add pgvector integration for PostgreSQL vector search
- Cache hot embeddings in Redis

**Priority:** ⭐⭐ Cost savings depend on cache hit rate

---

### 8. Memory Service Evolution Logic
**Location:** `services/memory/app/notes.py`

**Current State:** Basic CRUD, evolution endpoint exists but logic incomplete

**Agent Task:** Implement A-MEM evolution:
- Automatic importance decay over time
- Content summarization when merging related notes
- Link creation based on semantic similarity
- Pruning strategy based on access patterns

**Priority:** ⭐⭐ Memory quality over time

---

### 9. Billing Stripe Integration
**Location:** `services/billing/app/stripe.py`

**Current State:** Stripe service skeleton, not tested

**Agent Task:** Complete Stripe integration:
- Create test mode integration
- Implement webhook handlers (invoice.payment_succeeded, subscription.updated)
- Add usage-based billing with metered subscriptions
- Test with Stripe CLI

**Reference:** `agent_init/omni` Stripe implementation in SKILL.md

**Priority:** ⭐⭐ Monetization

---

### 10. Gateway Auth Production Implementation
**Location:** `services/gateway/app/auth/auth.go`

**Current State:** In-memory tenant store, JWT validation incomplete

**Agent Task:** Production auth:
- PostgreSQL-backed tenant store
- Proper JWT secret management from env
- API key hashing (bcrypt)
- Rate limit enforcement per tenant
- Quota checking before inference

**Priority:** ⭐⭐ Multi-tenancy security

---

## Lower Priority - Enhancement

### 11. Health Check Implementations
**Location:** All services

**Current State:** Health endpoints return static responses

**Agent Task:** Add real health checks:
- Database connectivity test
- Redis ping
- External service dependencies (gateway → Arkham, etc.)
- Return degraded status when dependencies fail

**Priority:** ⭐ Operational readiness

---

### 12. Observability Metrics
**Location:** `services/gateway/app/observability/observability.go`

**Current State:** Logging works, metrics are TODOs

**Agent Task:** Implement OpenTelemetry metrics:
- Request duration histogram
- Request count by endpoint/status
- Token usage tracking
- Cache hit/miss rates
- Model escalation events
- Dashboard queries (Grafana-compatible)

**Priority:** ⭐ Production visibility

---

### 13. Docker Build Optimization
**Location:** All `Dockerfile`s

**Current State:** Basic single-stage builds

**Agent Task:** Multi-stage builds:
- Separate build and runtime stages
- Smaller final images (distroless or alpine)
- Layer caching for dependencies
- Build arguments for version injection

**Priority:** ⭐ Deployment efficiency

---

### 14. Test Coverage
**Location:** `services/*/tests/` (mostly empty)

**Current State:** Minimal to no tests

**Agent Task:** Add test coverage:
- Unit tests for core logic (detectors, parsers, routers)
- Integration tests for service boundaries
- End-to-end workflow tests
- Load testing for gateway

**Priority:** ⭐⭐ Production confidence

---

### 15. Next.js Web Frontend
**Location:** `apps/web/` (empty)

**Current State:** Directory structure only

**Agent Task:** Build BIM frontend:
- Project upload page (IFC files)
- Project status dashboard
- Issue viewer
- Search across BIM elements
- Integration with gateway API

**Reference:** `agent_init/fullstack-bim/src/` for existing patterns

**Priority:** ⭐⭐ User-facing proof

---

## Agent Execution Notes

### Patterns to Follow

1. **Read before writing:** Always check `agent_init/` repos for existing implementations before creating new code

2. **Graphify after changes:** Run `graphify update .` after adding significant code so the knowledge graph stays current

3. **Test with real data:** Use `agent_init/BIMS_Structural.ifc` for BIM testing, not just mock data

4. **Follow the cost ladder:** When implementing model routing, default to cheapest option first

5. **Preserve state separation:** Workflow state, operational state, and cognitive state must remain separate in orchestration

### Common Pitfalls to Avoid

- ❌ Don't implement direct model calls in vertical apps (violates non-negotiable rule #1)
- ❌ Don't hardcode premium models as default (violates rule #3)
- ❌ Don't skip idempotency in side-effecting operations (violates rule #4)
- ❌ Don't create unscoped memory (violates rule #5)
- ❌ Don't allow cross-tenant data leakage in cache/memory queries

### Quick Wins for Agents

1. **Add missing `__init__.py` files** in Python packages
2. **Create `.env.example` files** for each service with required environment variables
3. **Add Makefiles** for common dev operations (build, test, run)
4. **Generate OpenAPI specs** from FastAPI services
5. **Create Postman/Insomnia collections** for API testing

---

## Current Build Status

| Component | Status | Agent Priority |
|-----------|--------|----------------|
| Gateway skeleton | ✅ Complete | - |
| Gateway model routing | ✅ Complete | - |
| Gateway ↔ Arkham integration | ✅ Complete | - |
| Gateway ↔ Provider integration | ✅ Complete | - |
| BIM ingestion | 🟡 Mock parser works | ⭐⭐⭐ |
| Orchestration flows | ✅ Complete | - |
| Orchestration task executors | ✅ Complete | - |
| Memory service | ✅ Complete | - |
| Semantic cache | ✅ Complete | - |
| Billing | ✅ Complete | - |
| Arkham detector | ✅ Complete | - |
| Arkham deception | ✅ Complete | - |
| Arkham fingerprint | ✅ Complete | - |
| Arkham audit | ✅ Complete | - |
| Web frontend | ✅ Complete | - |

**Legend:** ✅ Complete | 🟡 Partial | 🔴 Not started

**Knowledge Graph:** 743 nodes, 1,401 edges, 19 communities

---

*Last updated: Phase 1 Foundation complete - All core services implemented*

---

*Last updated: Phase 1 Foundation complete*
