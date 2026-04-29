# Graph Report - FullStackArkham  (2026-04-27)

## Corpus Check
- 64 files · ~31,543 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 846 nodes · 1578 edges · 26 communities detected
- Extraction: 59% EXTRACTED · 41% INFERRED · 0% AMBIGUOUS · INFERRED: 640 edges (avg confidence: 0.6)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 31|Community 31]]

## God Nodes (most connected - your core abstractions)
1. `AttackerFingerprinter` - 48 edges
2. `IFCParser` - 46 edges
3. `ThreatDetector` - 46 edges
4. `CacheStorage` - 38 edges
5. `lifespan()` - 36 edges
6. `MemoryNoteStore` - 32 edges
7. `DeceptionGenerator` - 30 edges
8. `PlanManager` - 30 edges
9. `BIMStorage` - 29 edges
10. `SecurityAuditor` - 29 edges

## Surprising Connections (you probably didn't know these)
- `run()` --calls--> `NewDatabase()`  [INFERRED]
  services/gateway/app/main.go → services/gateway/app/database/database.go
- `run()` --calls--> `NewTenantStore()`  [INFERRED]
  services/gateway/app/main.go → services/gateway/app/database/tenant_store.go
- `run()` --calls--> `contextKey`  [INFERRED]
  services/gateway/app/main.go → services/gateway/app/auth/auth.go
- `MemoryLinker` --uses--> `BIM Ingestion Service  Handles IFC file parsing, normalization, and storage for`  [INFERRED]
  services/memory/app/links.py → services/bim_ingestion/app/main.py
- `MemoryLinker` --uses--> `Request to check/store in cache`  [INFERRED]
  services/memory/app/links.py → services/semantic-cache/app/main.py

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (95): Security Auditor  Audit logging and fingerprint tracking for Arkham security eve, Calculate threat level based on attack count, Get security events for a tenant, Get security statistics for a tenant, PostgreSQL-backed security audit logging, Get information about an attacker fingerprint, Initialize database connection pool, Close database connections (+87 more)

### Community 1 - "Community 1"
Cohesion: 0.05
Nodes (52): ABC, ArtifactStorageExecutor, Artifact Storage Task Executor  Store workflow outputs to object storage., Store to object storage (S3, GCS, etc.)                  In production, would us, Store to database                  In production, would insert into artifacts ta, Get the name of the previous step, Store workflow artifacts to object storage, Store workflow output as artifact                  Step data should contain: (+44 more)

### Community 2 - "Community 2"
Cohesion: 0.04
Nodes (56): BillingRecord, calculate_cost(), create_invoice(), get_billing(), get_plan(), get_usage(), health_check(), list_plans() (+48 more)

### Community 3 - "Community 3"
Cohesion: 0.04
Nodes (55): IFCParser, IFC File Parser  Parses IFC files using IfcOpenShell and extracts: - Building el, Extract building elements from IFC file, Extract properties from an element, Extract quantities from an element, Extract materials from an element, Parser for IFC files using IfcOpenShell, Get the spatial container of an element (+47 more)

### Community 4 - "Community 4"
Cohesion: 0.04
Nodes (52): CheckpointStore, Checkpoint Store  Persistent storage for workflow state and checkpoints. Enables, Get workflows for a tenant, Save a workflow step checkpoint, Get failed workflows for retry processing, PostgreSQL-backed checkpoint storage, Initialize database connection pool, Close database connections (+44 more)

### Community 5 - "Community 5"
Cohesion: 0.05
Nodes (51): BaseModel, MemoryLinker, Memory Linker  Creates and manages links between memory notes. Links enable trav, Remove links to notes that no longer exist, Creates links between memory notes, Create links between notes                  Links are based on:         - Shared, Suggest the type of relationship between two notes, create_note() (+43 more)

### Community 6 - "Community 6"
Cohesion: 0.04
Nodes (55): NewAnthropicProvider(), NewGoogleProvider(), checkSemanticCache(), executeInference(), InferenceHandler(), recordUsage(), shouldCache(), storeInCache() (+47 more)

### Community 7 - "Community 7"
Cohesion: 0.05
Nodes (44): EmbeddingService, Embedding Service  Generates embeddings for semantic cache lookup. Uses sentence, Local embedding generation using sentence transformers, Lazy initialization of the model, Generate embedding for text                  Args:             text: Text to emb, Generate embeddings for multiple texts, Generate a deterministic mock embedding, Execute model inference                  Step data should contain:         - mod (+36 more)

### Community 8 - "Community 8"
Cohesion: 0.05
Nodes (24): BlockResponse, Client, DeceptionResponse, ThreatClassification, Claims, contextKey, InMemoryTenantStore, NewInMemoryTenantStore() (+16 more)

### Community 9 - "Community 9"
Cohesion: 0.06
Nodes (17): bim_file(), client(), End-to-End Test: BIM Ingestion Workflow  Tests the complete workflow: 1. Upload, Test orchestration workflow execution, Test gateway inference with model routing, Test Arkham security integration, Complete end-to-end workflow test, Create async HTTP client (+9 more)

### Community 10 - "Community 10"
Cohesion: 0.1
Nodes (21): BaseSettings, HealthHandler(), HealthResponse, ReadyHandler(), Status, responseWriter, ArkhamConfig, AuthConfig (+13 more)

### Community 11 - "Community 11"
Cohesion: 0.12
Nodes (9): APIError, APIKey, Tenant, TenantStore, Config, New(), parseLogLevel(), extractKeyPrefix() (+1 more)

### Community 13 - "Community 13"
Cohesion: 1.0
Nodes (1): Test gateway health endpoint

### Community 14 - "Community 14"
Cohesion: 1.0
Nodes (1): Test BIM ingestion service health

### Community 15 - "Community 15"
Cohesion: 1.0
Nodes (1): Test orchestration service health

### Community 16 - "Community 16"
Cohesion: 1.0
Nodes (1): Test uploading an IFC file

### Community 17 - "Community 17"
Cohesion: 1.0
Nodes (1): Test retrieving project elements

### Community 18 - "Community 18"
Cohesion: 1.0
Nodes (1): Test starting a BIM analysis workflow

### Community 19 - "Community 19"
Cohesion: 1.0
Nodes (1): Test getting workflow status

### Community 20 - "Community 20"
Cohesion: 1.0
Nodes (1): Test listing available workflow types

### Community 21 - "Community 21"
Cohesion: 1.0
Nodes (1): Test inference request through gateway

### Community 22 - "Community 22"
Cohesion: 1.0
Nodes (1): Test that model routing headers are present

### Community 23 - "Community 23"
Cohesion: 1.0
Nodes (1): Test that benign requests pass through

### Community 24 - "Community 24"
Cohesion: 1.0
Nodes (1): Test that scanner patterns are detected

### Community 25 - "Community 25"
Cohesion: 1.0
Nodes (1): Test complete BIM workflow:         1. Upload IFC file         2. Start analysis

### Community 31 - "Community 31"
Cohesion: 1.0
Nodes (1): Lazy initialization of Stripe client

## Knowledge Gaps
- **233 isolated node(s):** `End-to-End Test: BIM Ingestion Workflow  Tests the complete workflow: 1. Upload`, `Create async HTTP client`, `Get path to test BIM file`, `Test service health endpoints`, `Test gateway health endpoint` (+228 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 13`** (1 nodes): `Test gateway health endpoint`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 14`** (1 nodes): `Test BIM ingestion service health`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 15`** (1 nodes): `Test orchestration service health`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 16`** (1 nodes): `Test uploading an IFC file`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 17`** (1 nodes): `Test retrieving project elements`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 18`** (1 nodes): `Test starting a BIM analysis workflow`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 19`** (1 nodes): `Test getting workflow status`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 20`** (1 nodes): `Test listing available workflow types`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 21`** (1 nodes): `Test inference request through gateway`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 22`** (1 nodes): `Test that model routing headers are present`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 23`** (1 nodes): `Test that benign requests pass through`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 24`** (1 nodes): `Test that scanner patterns are detected`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 25`** (1 nodes): `Test complete BIM workflow:         1. Upload IFC file         2. Start analysis`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 31`** (1 nodes): `Lazy initialization of Stripe client`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `lifespan()` connect `Community 5` to `Community 0`, `Community 2`, `Community 3`, `Community 4`, `Community 7`, `Community 8`?**
  _High betweenness centrality (0.303) - this node is a cross-community bridge._
- **Why does `run()` connect `Community 6` to `Community 8`, `Community 3`, `Community 10`, `Community 11`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **Why does `BIM Ingestion Service  Handles IFC file parsing, normalization, and storage for` connect `Community 4` to `Community 0`, `Community 2`, `Community 3`, `Community 5`, `Community 7`?**
  _High betweenness centrality (0.128) - this node is a cross-community bridge._
- **Are the 36 inferred relationships involving `AttackerFingerprinter` (e.g. with `ThreatClassification` and `DeceptionResponse`) actually correct?**
  _`AttackerFingerprinter` has 36 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `IFCParser` (e.g. with `Background Jobs for BIM Ingestion  Handles heavy parsing tasks asynchronously vi` and `Background job to parse IFC files          This job is meant to be run asynchron`) actually correct?**
  _`IFCParser` has 34 INFERRED edges - model-reasoned connections that need verification._
- **Are the 36 inferred relationships involving `ThreatDetector` (e.g. with `ThreatClassification` and `DeceptionResponse`) actually correct?**
  _`ThreatDetector` has 36 INFERRED edges - model-reasoned connections that need verification._
- **Are the 24 inferred relationships involving `CacheStorage` (e.g. with `CacheLookup` and `Cache Lookup  Redis + PostgreSQL hybrid cache lookup for low-latency retrieval.`) actually correct?**
  _`CacheStorage` has 24 INFERRED edges - model-reasoned connections that need verification._