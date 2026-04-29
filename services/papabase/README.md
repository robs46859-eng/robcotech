# Papabase - Family Studio Suite

**Papabase** is a vertical application built on top of the stack-arkham core platform. It provides a family studio suite for business operations with AI-powered website generation ("Dad AI").

## Overview

Papabase combines:
- **Dad AI** - AI-powered website generation (Develop Another Day - Artificial Intelligence)
- **CRM** - Lead management, task tracking, workflow automation
- **Billing** - Subscription management, usage tracking, invoicing
- **Web Frontend** - Generated websites from single-page HTML to complex dashboards

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Papabase Service                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Dad AI    │  │    CRM      │  │     Billing         │  │
│  │   Agent     │  │   (Leads/   │  │   (Usage/Plans/     │  │
│  │             │  │    Tasks)   │  │    Invoices)        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    API Handlers (Go/HTTP)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   stack-arkham Core                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Gateway   │  │   Arkham    │  │     Memory          │  │
│  │  (Model     │  │  (Security/ │  │   (A-MEM evolving   │  │
│  │   Routing)  │  │  Fingerprint│  │    memory)          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Semantic   │  │    BIM      │  │   Orchestration     │  │
│  │   Cache     │  │  Ingestion  │  │   (Workflows)       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Pricing Tiers

| Tier | Price | Seats | AI Generations | Output Types |
|------|-------|-------|----------------|--------------|
| **Starter** | $29/mo | 1 | 3/mo | Single-page HTML |
| **Studio** | $99/mo | 5 | 15/mo | Multi-page React |
| **Agency** | $299/mo | 20 | Unlimited | Dashboards + Apps |
| **Enterprise** | Custom | Unlimited | Unlimited | Custom |

## Dad AI Output by Tier

### Starter ($29/mo)
- Single-page HTML websites
- Responsive design, SEO optimized
- Contact forms
- Deploy to papabase.subdomain.com

### Studio ($99/mo)
- Multi-page React sites (up to 10 pages)
- Blog integration, booking system
- Custom domain support
- Lead capture → CRM auto-import
- Analytics dashboard

### Agency ($299/mo)
- Full web applications
- Multi-user dashboards
- Client portals
- Payment integration
- API access
- White-label option

## Quick Start

### Run with Docker Compose

```bash
# Start all services including Papabase
docker compose up -d

# Check Papabase status
docker compose logs papabase

# Access Papabase API
curl http://localhost:8087/health
```

### Run Locally

```bash
cd services/papabase

# Install dependencies
go mod download

# Run the service
go run ./app --port 8087 --gateway-url http://localhost:8080
```

## API Endpoints

### Health
- `GET /health` - Service health check
- `GET /ready` - Readiness probe

### CRM - Leads
- `POST /api/v1/leads` - Create a new lead
- `GET /api/v1/leads` - List all leads
- `GET /api/v1/leads/{id}` - Get a specific lead
- `PUT /api/v1/leads/{id}` - Update a lead
- `DELETE /api/v1/leads/{id}` - Delete a lead

### CRM - Tasks
- `POST /api/v1/tasks` - Create a new task
- `GET /api/v1/tasks` - List all tasks
- `GET /api/v1/tasks/{id}` - Get a specific task
- `PUT /api/v1/tasks/{id}` - Update a task
- `DELETE /api/v1/tasks/{id}` - Delete a task

### Dad AI - Website Generation
- `POST /api/v1/ai/generate` - Generate a new website
- `GET /api/v1/ai/templates` - List available templates
- `GET /api/v1/ai/projects` - List website projects
- `GET /api/v1/ai/projects/{id}` - Get a specific project

### Pricing & Billing
- `GET /api/v1/pricing/plans` - List subscription plans
- `GET /api/v1/billing/usage` - Get usage data
- `GET /api/v1/billing/invoices` - List invoices

## Example API Calls

### Create a Lead
```bash
curl -X POST http://localhost:8087/api/v1/leads \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1-555-123-4567",
    "company": "Acme Corp",
    "source": "website",
    "tenant_id": "default"
  }'
```

### Generate a Website
```bash
curl -X POST http://localhost:8087/api/v1/ai/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a professional website for a plumbing business",
    "business_type": "Plumbing Services",
    "tier": "studio",
    "output_type": "multi_page",
    "color_scheme": "blue",
    "features": ["contact_form", "service_list", "testimonials"]
  }'
```

### List Pricing Plans
```bash
curl http://localhost:8087/api/v1/pricing/plans
```

## Development

### Project Structure

```
services/papabase/
├── app/
│   ├── main.go              # Service entry point
│   ├── api/
│   │   ├── handlers.go      # CRM & Task handlers
│   │   └── dad_ai_handlers.go # Dad AI & Billing handlers
│   └── agents/
│       └── dad_ai.go        # Dad AI agent implementation
├── cmd/                      # CLI commands (future)
├── templates/                # Website templates (future)
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies
├── Dockerfile               # Container build
└── pricing-tiers.yaml       # Pricing configuration
```

### Adding New Features

1. **New CRM fields**: Update the `Lead` or `Task` struct in `handlers.go`
2. **New AI templates**: Add to `ListTemplatesHandler` in `dad_ai_handlers.go`
3. **New pricing tier**: Update `pricing-tiers.yaml` and `ListPlansHandler`
4. **New Dad AI capabilities**: Extend the `DadAIAgent` in `dad_ai.go`

## Integration with stack-arkham

Papabase uses the stack-arkham gateway for all LLM inference:

```go
// Dad AI calls the gateway for website generation
response, err := dadAI.callGateway(ctx, prompt)
```

The gateway handles:
- Model routing (cost ladder)
- Semantic caching
- Arkham security
- Usage tracking for billing

## License

AGPL v3 - See [LICENSE](../../LICENSE)
