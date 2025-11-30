# Full-Stack Web Application Design Document
## MCP-Docs Platform

**Version:** 1.0  
**Date:** 2024  
**Status:** Design Phase

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [Architecture](#architecture)
4. [Technology Stack](#technology-stack)
5. [Component Design](#component-design)
6. [UI Design & User Experience](#ui-design--user-experience)
7. [Data Models](#data-models)
8. [API Design](#api-design)
9. [Message Flow & Kafka Integration](#message-flow--kafka-integration)
10. [Agent System & MCP Server Management](#agent-system--mcp-server-management)
11. [Database Schema](#database-schema)
12. [Security & Authentication](#security--authentication)
13. [Deployment Architecture](#deployment-architecture)
14. [Implementation Phases](#implementation-phases)
15. [Scalability Considerations](#scalability-considerations)

---

## Executive Summary

This document outlines the design for converting the MCP-Docs CLI tool into a full-stack web application with a **subscription-based SaaS model**. The platform will enable users to manage documentation indexing projects through a web interface, with asynchronous processing, real-time updates, and scalable agent-based MCP server management. **All API keys are managed by the platform** - users subscribe to tiers based on usage (SSE executions).

### Key Features
- **Subscription-based SaaS model** with tiered access
- Web-based project management UI
- Asynchronous documentation indexing
- Real-time progress tracking
- **Usage-based billing** (tracked via SSE executions)
- **Advanced AI features** (CrewAI integration) for premium tiers
- **Enhanced MCP tools** (search_docs with AI, find_code_snippet) for advanced tier
- Agent-based MCP server orchestration
- Hosted vector database integration
- PostgreSQL for metadata and configuration
- Platform-managed embedding API keys

---

## System Overview

### Current State (CLI)
- Python-based CLI tool
- Local ChromaDB storage
- Synchronous processing
- Single-user, local execution
- Manual MCP server management

### Target State (Web Platform)
- Next.js frontend with modern UI
- Golang backend API
- Kafka for async job processing
- Hosted vector database (Pinecone/Weaviate/Qdrant)
- PostgreSQL for metadata
- Agent workers for MCP server lifecycle
- Multi-user, cloud-ready architecture
- **Subscription-based access model**
- **Platform-managed API keys** (no user keys required)
- **Usage tracking and billing** (SSE execution-based)
- **Tiered feature access** (Free, Basic, Pro, Advanced)
- **CrewAI-powered AI features** for advanced tier

---

## Architecture

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend Layer                           │
│                         (Next.js)                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Dashboard  │  │  Project Mgmt│  │   Search UI  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP/REST + WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway Layer                          │
│                        (Golang)                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  REST API    │  │  WebSocket   │  │  Auth Service│          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┼─────────────┐
                │             │             │
                ▼             ▼             ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │PostgreSQL│  │  Kafka   │  │  Vector  │
        │          │  │          │  │    DB    │
        └──────────┘  └──────────┘  └──────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │         Worker Agents (Python)           │
        │  ┌──────────────┐  ┌──────────────┐     │
        │  │ Index Worker │  │ MCP Agent    │     │
        │  └──────────────┘  └──────────────┘     │
        └─────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │   MCP Servers    │
                    │  (Per Project)   │
                    └──────────────────┘
```

### Architecture Layers

1. **Presentation Layer (Next.js)**
   - User interface and interactions
   - Real-time updates via WebSocket
   - Client-side state management

2. **API Layer (Golang)**
   - RESTful API endpoints
   - WebSocket server for real-time updates
   - Authentication & authorization
   - Request validation and routing

3. **Message Queue (Kafka)**
   - Job queue for indexing tasks
   - Event streaming for status updates
   - Decoupling of services

4. **Data Layer**
   - **PostgreSQL**: User data, projects, configurations, job status
   - **Vector DB**: Document embeddings and semantic search

5. **Worker Layer (Python Agents)**
   - Indexing workers (scraping, embedding, storage)
   - MCP server agents (lifecycle management)

---

## Technology Stack

### Frontend
- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript
- **UI Library**: Tailwind CSS + shadcn/ui or Material-UI
- **State Management**: Zustand or React Query
- **Real-time**: WebSocket client (native or Socket.io client)
- **Forms**: React Hook Form + Zod validation

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Gin or Echo
- **ORM**: GORM or sqlx
- **WebSocket**: Gorilla WebSocket or native net/http
- **Message Queue**: Confluent Kafka Go client
- **Vector DB Client**: Provider-specific SDK (Pinecone/Weaviate/Qdrant)

### Infrastructure
- **Message Queue**: Apache Kafka (Confluent Cloud or self-hosted)
- **Database**: PostgreSQL 15+ (hosted: AWS RDS, Supabase, or Neon)
- **Vector Database**: 
  - **Option 1**: Pinecone (managed, recommended for production)
  - **Option 2**: Weaviate Cloud
  - **Option 3**: Qdrant Cloud
- **Containerization**: Docker + Docker Compose (dev), Kubernetes (prod)
- **CI/CD**: GitHub Actions or GitLab CI

### Worker Agents
- **Language**: Python 3.10+ (existing codebase)
- **Framework**: Celery (optional) or direct Kafka consumer
- **Scraping**: Playwright (existing)
- **Embeddings**: OpenAI SDK / Azure OpenAI SDK (existing)
- **MCP Server**: FastMCP (existing)
- **AI Framework**: CrewAI (for Advanced tier features)
- **LLM**: OpenAI GPT-4 (for CrewAI agents)

---

## Component Design

### 1. Frontend Components (Next.js)

#### Pages
- `/` - Dashboard (project overview, recent activity, usage stats)
- `/projects` - Project list and management
- `/projects/[id]` - Project details and configuration
- `/projects/[id]/index` - Indexing status and logs
- `/projects/[id]/search` - Search interface for indexed docs
- `/subscription` - Subscription management and billing
- `/usage` - Usage dashboard (SSE executions, tier limits)
- `/settings` - User settings
- `/auth/login` - Authentication
- `/auth/register` - User registration

#### Key Components
- `ProjectCard` - Project summary card
- `IndexingProgress` - Real-time indexing progress bar
- `SearchInterface` - Semantic search UI with results (tier-aware)
- `MCPStatusBadge` - MCP server status indicator
- `LogViewer` - Streaming log viewer for indexing jobs
- `ProjectForm` - Create/edit project form
- `SubscriptionTierCard` - Display current tier and features
- `UsageMeter` - Real-time usage tracking (SSE executions)
- `TierUpgradePrompt` - Upgrade prompts for premium features
- `AISearchInterface` - CrewAI-powered search (Advanced tier only)

### 2. Backend Services (Golang)

#### Core Services

**Project Service**
- `CreateProject(userID, name, url, config) -> Project`
- `GetProject(projectID) -> Project`
- `ListProjects(userID) -> []Project`
- `UpdateProject(projectID, updates) -> Project`
- `DeleteProject(projectID) -> error`

**Indexing Service**
- `StartIndexing(projectID, options) -> Job`
- `GetIndexingStatus(jobID) -> JobStatus`
- `CancelIndexing(jobID) -> error`
- `GetIndexingLogs(jobID) -> []LogEntry`

**Search Service**
- `SearchDocuments(projectID, query, topK) -> []SearchResult`
- `GetDocument(projectID, docID) -> Document`

**MCP Agent Service**
- `StartMCPServer(projectID) -> MCPInstance`
- `StopMCPServer(projectID) -> error`
- `GetMCPStatus(projectID) -> MCPStatus`
- `ListMCPInstances(userID) -> []MCPInstance`

**User Service**
- `RegisterUser(email, password) -> User`
- `AuthenticateUser(email, password) -> Token`
- `GetUser(userID) -> User`
- `UpdateUserSettings(userID, settings) -> User`

**Subscription Service**
- `GetSubscription(userID) -> Subscription`
- `UpgradeSubscription(userID, tier) -> Subscription`
- `CancelSubscription(userID) -> error`
- `GetUsageStats(userID, period) -> UsageStats`
- `CheckFeatureAccess(userID, feature) -> bool`
- `TrackSSEExecution(userID, projectID) -> error`

**Billing Service**
- `GetBillingHistory(userID) -> []Invoice`
- `GetCurrentUsage(userID) -> Usage`
- `CheckUsageLimit(userID) -> (remaining, limit)`

### 3. Worker Agents (Python)

#### Index Worker
- Consumes Kafka messages for indexing jobs
- Executes scraping pipeline (existing code)
- Generates embeddings (existing code)
- Stores in vector database
- Publishes progress updates to Kafka
- Updates PostgreSQL job status

#### MCP Agent
- Manages MCP server lifecycle
- Starts/stops MCP servers per project
- Monitors MCP server health
- Publishes status updates
- Handles server restarts and failures

---

## Data Models

### PostgreSQL Schema

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    subscription_tier VARCHAR(50) DEFAULT 'free', -- free, basic, pro, advanced
    subscription_status VARCHAR(50) DEFAULT 'active', -- active, cancelled, expired, trial
    subscription_started_at TIMESTAMP,
    subscription_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_subscription_tier (subscription_tier),
    INDEX idx_subscription_status (subscription_status)
);
```

#### Projects Table
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    -- Embedding provider is platform-managed, no user config needed
    vector_db_collection_id VARCHAR(255), -- Reference to vector DB collection
    status VARCHAR(50) DEFAULT 'pending', -- pending, indexing, completed, failed
    indexed_pages_count INTEGER DEFAULT 0,
    total_pages_count INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
);
```

#### Indexing Jobs Table
```sql
CREATE TABLE indexing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'queued', -- queued, running, completed, failed, cancelled
    max_pages INTEGER,
    max_depth INTEGER,
    pages_scraped INTEGER DEFAULT 0,
    pages_indexed INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_project_id (project_id),
    INDEX idx_status (status)
);
```

#### Job Logs Table
```sql
CREATE TABLE job_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES indexing_jobs(id) ON DELETE CASCADE,
    level VARCHAR(20), -- info, warning, error
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_job_id (job_id),
    INDEX idx_created_at (created_at)
);
```

#### MCP Instances Table
```sql
CREATE TABLE mcp_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'stopped', -- stopped, starting, running, error
    port INTEGER,
    process_id INTEGER,
    endpoint_url TEXT,
    error_message TEXT,
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_project_id (project_id),
    INDEX idx_status (status)
);
```

#### Subscriptions Table
```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(50) NOT NULL, -- free, basic, pro, advanced
    status VARCHAR(50) DEFAULT 'active', -- active, cancelled, expired, trial
    billing_cycle VARCHAR(20), -- monthly, yearly
    price DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    started_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    stripe_subscription_id VARCHAR(255), -- If using Stripe
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_tier (tier),
    INDEX idx_status (status)
);
```

#### Usage Tracking Table (SSE Executions)
```sql
CREATE TABLE usage_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL, -- 'sse_execution', 'indexing', 'search'
    tier_at_time VARCHAR(50), -- Tier when event occurred
    metadata JSONB, -- Additional event data
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at)
);

-- Monthly usage summary for billing
CREATE TABLE monthly_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL, -- 1-12
    sse_executions INTEGER DEFAULT 0,
    indexing_jobs INTEGER DEFAULT 0,
    searches INTEGER DEFAULT 0,
    tier VARCHAR(50), -- Tier during this month
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, year, month),
    INDEX idx_user_id (user_id),
    INDEX idx_year_month (year, month)
);
```

#### Platform API Keys Table (Encrypted - Internal)
```sql
-- Platform-managed API keys for embedding providers
-- These are shared across users based on subscription tier
CREATE TABLE platform_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL, -- 'openai' or 'azure-openai'
    key_name VARCHAR(255),
    encrypted_key TEXT NOT NULL, -- Encrypted API key
    config JSONB, -- Additional provider config
    tier_access TEXT[], -- Which tiers can use this key ['free', 'basic', 'pro', 'advanced']
    rate_limit_per_minute INTEGER,
    current_usage_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_provider (provider),
    INDEX idx_is_active (is_active)
);
```

### Vector Database Schema

#### Collection Structure
- **Collection Name**: `project_{project_id}` or `{user_id}_{project_name}`
- **Metadata Fields**:
  - `project_id`: UUID reference
  - `url`: Source URL
  - `title`: Page title
  - `chunk_index`: Integer (for multi-chunk pages)
  - `scraped_at`: Timestamp
  - `job_id`: UUID reference to indexing job

---

## API Design

### REST API Endpoints

#### Authentication
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/me
```

#### Projects
```
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/:id
PATCH  /api/v1/projects/:id
DELETE /api/v1/projects/:id
```

#### Indexing
```
POST   /api/v1/projects/:id/index
GET    /api/v1/projects/:id/index/status
GET    /api/v1/projects/:id/index/logs
POST   /api/v1/projects/:id/index/cancel
```

#### Search
```
POST   /api/v1/projects/:id/search
GET    /api/v1/projects/:id/documents/:docId
```

#### MCP Servers
```
POST   /api/v1/projects/:id/mcp/start
POST   /api/v1/projects/:id/mcp/stop
GET    /api/v1/projects/:id/mcp/status
GET    /api/v1/mcp/instances
```

#### Subscription & Usage
```
GET    /api/v1/subscription
POST   /api/v1/subscription/upgrade
POST   /api/v1/subscription/cancel
GET    /api/v1/usage/current
GET    /api/v1/usage/history
GET    /api/v1/usage/limits
POST   /api/v1/usage/track  -- Internal: Track SSE execution
```

#### Settings
```
GET    /api/v1/settings
PATCH  /api/v1/settings
```

### WebSocket Events

#### Client → Server
- `subscribe:project:{projectId}` - Subscribe to project updates
- `subscribe:job:{jobId}` - Subscribe to job updates
- `unsubscribe:project:{projectId}`
- `unsubscribe:job:{jobId}`

#### Server → Client
- `project:updated` - Project status changed
- `job:progress` - Indexing progress update
- `job:completed` - Indexing job completed
- `job:failed` - Indexing job failed
- `mcp:status` - MCP server status update
- `log:entry` - New log entry

### Request/Response Examples

#### Create Project
```http
POST /api/v1/projects
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "React Docs",
  "url": "https://react.dev",
  "description": "React documentation"
}

Response: 201 Created
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "React Docs",
  "url": "https://react.dev",
  "status": "pending",
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### Start Indexing
```http
POST /api/v1/projects/550e8400-e29b-41d4-a716-446655440000/index
Content-Type: application/json
Authorization: Bearer <token>

{
  "max_pages": 200,
  "max_depth": 5
}

Response: 202 Accepted
{
  "job_id": "660e8400-e29b-41d4-a716-446655440001",
  "status": "queued",
  "created_at": "2024-01-15T10:05:00Z"
}
```

#### Search Documents
```http
POST /api/v1/projects/550e8400-e29b-41d4-a716-446655440000/search
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "how to use hooks",
  "top_k": 5,
  "use_ai": false  -- Set to true for Advanced tier (CrewAI)
}

Response: 200 OK
{
  "results": [
    {
      "id": "doc_123",
      "score": 0.92,
      "document": "React Hooks are functions that let you...",
      "metadata": {
        "url": "https://react.dev/hooks",
        "title": "Introducing Hooks"
      }
    }
  ],
  "usage_tracked": true
}
```

#### Get Subscription
```http
GET /api/v1/subscription
Authorization: Bearer <token>

Response: 200 OK
{
  "tier": "pro",
  "status": "active",
  "billing_cycle": "monthly",
  "price": 29.99,
  "currency": "USD",
  "expires_at": "2024-02-15T10:00:00Z",
  "features": {
    "max_projects": 10,
    "max_sse_executions_per_month": 10000,
    "ai_search": true,
    "advanced_tools": true
  }
}
```

#### Get Usage
```http
GET /api/v1/usage/current
Authorization: Bearer <token>

Response: 200 OK
{
  "tier": "pro",
  "current_period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "usage": {
    "sse_executions": 3420,
    "indexing_jobs": 15,
    "searches": 892
  },
  "limits": {
    "sse_executions": 10000,
    "indexing_jobs": -1,  -- unlimited
    "searches": -1  -- unlimited
  },
  "remaining": {
    "sse_executions": 6580
  }
}
```

#### Upgrade Subscription
```http
POST /api/v1/subscription/upgrade
Content-Type: application/json
Authorization: Bearer <token>

{
  "tier": "advanced",
  "billing_cycle": "monthly"
}

Response: 200 OK
{
  "tier": "advanced",
  "status": "active",
  "checkout_url": "https://stripe.com/checkout/...",  -- If payment required
  "expires_at": "2024-02-15T10:00:00Z"
}
```

---

## Message Flow & Kafka Integration

### Kafka Topics

1. **`indexing-jobs`** - Indexing job requests
2. **`indexing-progress`** - Progress updates from workers
3. **`indexing-completed`** - Job completion events
4. **`mcp-commands`** - MCP server lifecycle commands
5. **`mcp-status`** - MCP server status updates

### Message Schemas

#### Indexing Job Request
```json
{
  "job_id": "uuid",
  "project_id": "uuid",
  "user_id": "uuid",
  "user_tier": "pro",
  "url": "string",
  "max_pages": 200,
  "max_depth": 5,
  "vector_db_collection_id": "string"
}
```

#### Progress Update
```json
{
  "job_id": "uuid",
  "project_id": "uuid",
  "pages_scraped": 50,
  "pages_indexed": 45,
  "current_url": "https://...",
  "status": "running"
}
```

#### Job Completed
```json
{
  "job_id": "uuid",
  "project_id": "uuid",
  "status": "completed",
  "total_pages": 200,
  "indexed_pages": 195,
  "errors": []
}
```

### Message Flow Diagrams

#### Indexing Flow
```
User → API → Kafka (indexing-jobs)
                ↓
         Index Worker (consumes)
                ↓
         Scraping → Embedding → Vector DB
                ↓
         Kafka (indexing-progress)
                ↓
         API → WebSocket → Frontend
                ↓
         PostgreSQL (update job status)
```

#### MCP Server Lifecycle
```
User → API → Kafka (mcp-commands)
                ↓
         MCP Agent (consumes)
                ↓
         Start/Stop MCP Server
                ↓
         Kafka (mcp-status)
                ↓
         API → WebSocket → Frontend
                ↓
         PostgreSQL (update instance status)
```

---

## Subscription Tiers & Features

### Tier Definitions

#### Free Tier
- **Price**: $0/month
- **SSE Executions**: 100/month
- **Projects**: 1
- **Indexing Jobs**: 1 concurrent
- **Max Pages per Index**: 50
- **Features**:
  - Basic semantic search
  - Standard MCP server with `search_docs` tool
  - Community support

#### Basic Tier
- **Price**: $9.99/month
- **SSE Executions**: 1,000/month
- **Projects**: 5
- **Indexing Jobs**: 3 concurrent
- **Max Pages per Index**: 200
- **Features**:
  - Basic semantic search
  - Standard MCP server with `search_docs` tool
  - Email support
  - Usage analytics

#### Pro Tier
- **Price**: $29.99/month
- **SSE Executions**: 10,000/month
- **Projects**: 20
- **Indexing Jobs**: 10 concurrent
- **Max Pages per Index**: 1,000
- **Features**:
  - Basic semantic search
  - Standard MCP server with `search_docs` tool
  - Priority support
  - Advanced usage analytics
  - API access

#### Advanced Tier
- **Price**: $99.99/month
- **SSE Executions**: 100,000/month
- **Projects**: Unlimited
- **Indexing Jobs**: Unlimited concurrent
- **Max Pages per Index**: Unlimited
- **Features**:
  - **AI-Powered Search** (CrewAI integration)
  - **Enhanced MCP Tools**:
    - `search_docs` - AI-enhanced semantic search with CrewAI
    - `find_code_snippet` - Intelligent code snippet finder
    - `explain_concept` - AI-powered concept explanation (future)
    - `generate_example` - Code example generation (future)
  - Priority support (24/7)
  - Advanced analytics and insights
  - API access with higher rate limits
  - Custom integrations

### Usage Tracking

**SSE Execution Definition**: Each time an MCP server tool is called via Server-Sent Events (SSE), it counts as one execution. This includes:
- `search_docs` calls
- `find_code_snippet` calls (Advanced tier)
- Any other tool invocations

**Tracking Mechanism**:
1. MCP server tracks each tool execution
2. Publishes usage event to Kafka
3. Backend service records in `usage_tracking` table
4. Monthly aggregation in `monthly_usage` table
5. Real-time usage limits enforced via API middleware

### Feature Access Control

- **Middleware**: Check user tier before allowing feature access
- **MCP Server**: Dynamically generate tools based on user tier
- **Frontend**: Show/hide features based on tier
- **Rate Limiting**: Enforce SSE execution limits per tier

---

## Agent System & MCP Server Management

### Agent Architecture

#### Index Worker Agent
- **Language**: Python
- **Responsibilities**:
  - Consume indexing jobs from Kafka
  - Execute scraping pipeline (existing `indexer` module)
  - Generate embeddings (existing `embeddings` module)
  - Store in vector database
  - Publish progress updates
  - Handle errors and retries

- **Implementation**:
```python
# Simplified structure
from kafka import KafkaConsumer, KafkaProducer
from indexer import index_documentation
from embeddings.openai_provider import OpenAIEmbeddingProvider

def process_indexing_job(message):
    job_data = json.loads(message.value)
    
    # Get platform-managed API key based on user tier
    api_key = get_platform_api_key(
        provider='openai',
        tier=job_data['user_tier']
    )
    
    # Initialize embedding provider with platform key
    provider = OpenAIEmbeddingProvider(api_key=api_key)
    
    # Initialize vector store (hosted)
    vector_store = get_vector_store(job_data['vector_db_collection_id'])
    
    # Index documentation
    collection = index_documentation(
        url=job_data['url'],
        provider=provider,
        max_pages=job_data['max_pages'],
        vector_store=vector_store
    )
    
    # Publish completion
    producer.send('indexing-completed', {
        'job_id': job_data['job_id'],
        'status': 'completed'
    })
```

#### MCP Agent
- **Language**: Python
- **Responsibilities**:
  - Manage MCP server lifecycle
  - Start MCP servers in isolated containers/processes
  - Monitor server health
  - Handle server restarts
  - Expose MCP servers via reverse proxy or direct ports
  - **Track SSE executions** for billing
  - **Generate tier-specific tools** based on user subscription

- **MCP Server Generation**:
  - Generate server.py dynamically (similar to existing CLI)
  - Configure with project-specific vector DB connection
  - **Inject user tier information** for tool generation
  - **Include tier-specific tools** (CrewAI, find_code_snippet for Advanced tier)
  - Use platform-managed API keys (not user keys)
  - Start as subprocess or container
  - Expose via port or internal service mesh
  - **Track usage** on each tool execution

- **Implementation Options**:
  1. **Subprocess Management** (Simple)
     - Start Python subprocess per MCP server
     - Manage via process IDs
     - Port allocation from pool
   
  2. **Container-based** (Scalable)
     - Docker container per MCP server
     - Kubernetes pods for production
     - Service discovery for routing

  3. **Serverless** (Advanced)
     - AWS Lambda / Google Cloud Functions
     - On-demand execution
     - Cost-effective for low usage

### MCP Server Configuration

Each MCP server needs:
- Vector database connection (hosted)
- **Platform-managed embedding API key** (based on user tier)
- Project-specific collection ID
- **User tier information** (for tool generation)
- Port allocation or endpoint URL
- **Usage tracking endpoint** (to record SSE executions)

### Tier-Based MCP Tools

#### Free, Basic, Pro Tiers
- **`search_docs`**: Standard semantic search
  - Vector similarity search
  - Returns top-k matching documents
  - No AI enhancement

#### Advanced Tier Only
- **`search_docs`**: AI-enhanced semantic search with CrewAI
  - Vector similarity search
  - **CrewAI agent** processes results
  - **Contextual understanding** and summarization
  - **Intelligent ranking** based on query intent
  - Returns enhanced results with AI insights

- **`find_code_snippet`**: Intelligent code snippet finder
  - Searches for code examples in documentation
  - **CrewAI agent** identifies relevant code patterns
  - Extracts and formats code snippets
  - Provides context and explanations
  - Returns multiple code examples with metadata

### CrewAI Integration (Advanced Tier)

**Architecture**:
```
User Query → Vector Search → Top Results
                ↓
         CrewAI Agent (Researcher)
                ↓
         Analyze & Enhance Results
                ↓
         CrewAI Agent (Writer)
                ↓
         Format & Summarize
                ↓
         Enhanced Response
```

**CrewAI Agents**:
1. **Researcher Agent**
   - Analyzes search results
   - Identifies most relevant information
   - Extracts key concepts
   - Determines query intent

2. **Writer Agent**
   - Summarizes findings
   - Formats response
   - Adds contextual explanations
   - Structures output

**Implementation**:
```python
from crewai import Agent, Task, Crew
from langchain_openai import ChatOpenAI

# CrewAI setup for Advanced tier search_docs
def enhanced_search_docs(query, vector_results, user_tier):
    if user_tier != 'advanced':
        return standard_search(vector_results)
    
    # Initialize CrewAI agents
    researcher = Agent(
        role='Research Analyst',
        goal='Analyze and understand the search results',
        backstory='Expert at analyzing documentation and extracting insights',
        llm=ChatOpenAI(model='gpt-4', temperature=0.3)
    )
    
    writer = Agent(
        role='Technical Writer',
        goal='Format and present findings clearly',
        backstory='Expert technical writer specializing in documentation',
        llm=ChatOpenAI(model='gpt-4', temperature=0.3)
    )
    
    # Create tasks
    research_task = Task(
        description=f'Analyze these search results for query: {query}',
        agent=researcher
    )
    
    write_task = Task(
        description='Format the analysis into a clear, helpful response',
        agent=writer
    )
    
    # Execute crew
    crew = Crew(agents=[researcher, writer], tasks=[research_task, write_task])
    result = crew.kickoff()
    
    return {
        'enhanced_results': result,
        'original_results': vector_results,
        'ai_enhanced': True
    }
```

**Code Snippet Finder**:
```python
def find_code_snippet(query, project_id, user_tier):
    if user_tier != 'advanced':
        raise PermissionError('Feature available in Advanced tier only')
    
    # Search for code-related content
    code_results = vector_search(
        query=f"{query} code example implementation",
        project_id=project_id,
        filter={'type': 'code'}
    )
    
    # CrewAI agent to extract and format code
    code_agent = Agent(
        role='Code Analyst',
        goal='Find and extract relevant code snippets',
        backstory='Expert at understanding code patterns and examples',
        llm=ChatOpenAI(model='gpt-4', temperature=0.2)
    )
    
    # Extract and format code snippets
    task = Task(
        description=f'Extract code snippets for: {query}',
        agent=code_agent
    )
    
    crew = Crew(agents=[code_agent], tasks=[task])
    result = crew.kickoff()
    
    return {
        'snippets': result.code_snippets,
        'explanations': result.context,
        'metadata': result.metadata
    }
```

### Agent Deployment

- **Development**: Docker Compose with Python workers
- **Production**: Kubernetes Deployment with:
  - Horizontal scaling
  - Health checks
  - Auto-restart on failure
  - Resource limits

---

## Database Schema

### Complete PostgreSQL Schema

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    email_verified BOOLEAN DEFAULT false,
    subscription_tier VARCHAR(50) DEFAULT 'free',
    subscription_status VARCHAR(50) DEFAULT 'active',
    subscription_started_at TIMESTAMP,
    subscription_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    billing_cycle VARCHAR(20),
    price DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    started_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    stripe_subscription_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Usage Tracking (SSE Executions)
CREATE TABLE usage_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    tier_at_time VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Monthly Usage Summary
CREATE TABLE monthly_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    sse_executions INTEGER DEFAULT 0,
    indexing_jobs INTEGER DEFAULT 0,
    searches INTEGER DEFAULT 0,
    tier VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, year, month)
);

-- Platform API Keys (Internal - Encrypted)
CREATE TABLE platform_api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(50) NOT NULL,
    key_name VARCHAR(255),
    encrypted_key TEXT NOT NULL,
    config JSONB,
    tier_access TEXT[],
    rate_limit_per_minute INTEGER,
    current_usage_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    vector_db_collection_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    indexed_pages_count INTEGER DEFAULT 0,
    total_pages_count INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexing Jobs
CREATE TABLE indexing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'queued',
    max_pages INTEGER,
    max_depth INTEGER,
    pages_scraped INTEGER DEFAULT 0,
    pages_indexed INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Job Logs
CREATE TABLE job_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID REFERENCES indexing_jobs(id) ON DELETE CASCADE,
    level VARCHAR(20),
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- MCP Instances
CREATE TABLE mcp_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'stopped',
    port INTEGER,
    process_id INTEGER,
    endpoint_url TEXT,
    error_message TEXT,
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_subscription_tier ON users(subscription_tier);
CREATE INDEX idx_users_subscription_status ON users(subscription_status);
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_indexing_jobs_project_id ON indexing_jobs(project_id);
CREATE INDEX idx_indexing_jobs_status ON indexing_jobs(status);
CREATE INDEX idx_job_logs_job_id ON job_logs(job_id);
CREATE INDEX idx_mcp_instances_project_id ON mcp_instances(project_id);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_tier ON subscriptions(tier);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_usage_tracking_user_id ON usage_tracking(user_id);
CREATE INDEX idx_usage_tracking_project_id ON usage_tracking(project_id);
CREATE INDEX idx_usage_tracking_event_type ON usage_tracking(event_type);
CREATE INDEX idx_usage_tracking_created_at ON usage_tracking(created_at);
CREATE INDEX idx_monthly_usage_user_id ON monthly_usage(user_id);
CREATE INDEX idx_monthly_usage_year_month ON monthly_usage(year, month);
CREATE INDEX idx_platform_api_keys_provider ON platform_api_keys(provider);
CREATE INDEX idx_platform_api_keys_is_active ON platform_api_keys(is_active);
```

---

## Security & Authentication

### Authentication Strategy

1. **JWT-based Authentication**
   - Access tokens (short-lived, 15 minutes)
   - Refresh tokens (long-lived, 7 days)
   - Token stored in httpOnly cookies or localStorage

2. **Password Hashing**
   - bcrypt with salt rounds (10+)
   - Never store plaintext passwords

3. **Platform API Key Encryption**
   - Encrypt platform API keys at rest (AES-256)
   - Use master encryption key (stored in secure vault)
   - Decrypt only when needed for API calls
   - Rotate keys periodically
   - Track usage per key for rate limiting

### Authorization

- **Project-level**: Users can only access their own projects
- **Resource-level**: Check ownership on all operations
- **Tier-based**: Enforce feature access based on subscription tier
- **Usage Limits**: Enforce SSE execution limits per tier
- **Role-based** (future): Admin, User, Viewer roles

### Usage Tracking & Rate Limiting

- **SSE Execution Tracking**: Every MCP tool call is tracked
- **Middleware**: Check usage limits before processing requests
- **Real-time Limits**: Enforce monthly SSE execution limits
- **Graceful Degradation**: Return appropriate errors when limits exceeded
- **Usage Analytics**: Track and display usage to users

### Security Best Practices

1. **Input Validation**
   - Validate all user inputs
   - Sanitize URLs before scraping
   - Rate limiting on API endpoints

2. **CORS Configuration**
   - Restrict to frontend domain
   - Credentials: true for cookies

3. **HTTPS Only**
   - Enforce HTTPS in production
   - Secure cookies (Secure, SameSite)

4. **Vector DB Security**
   - Use API keys for vector DB access
   - Isolate collections per user/project
   - Network security (VPC, private endpoints)

---

## Deployment Architecture

### Development Environment

```
docker-compose.yml:
  - PostgreSQL (local)
  - Kafka + Zookeeper (local)
  - Backend API (Golang, hot reload)
  - Frontend (Next.js, dev server)
  - Index Worker (Python)
  - MCP Agent (Python)
```

### Production Environment

#### Option 1: Cloud-Native (Recommended)
- **Frontend**: Vercel or Netlify
- **Backend**: AWS ECS / Google Cloud Run / Azure Container Apps
- **Database**: AWS RDS / Google Cloud SQL / Azure Database
- **Kafka**: Confluent Cloud
- **Vector DB**: Pinecone / Weaviate Cloud
- **Agents**: AWS ECS / Kubernetes

#### Option 2: Self-Hosted
- **Infrastructure**: Kubernetes cluster
- **Ingress**: Nginx or Traefik
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack or Loki

### Environment Variables

#### Backend (Golang)
```env
DATABASE_URL=postgresql://...
KAFKA_BROKERS=localhost:9092
VECTOR_DB_TYPE=pinecone
VECTOR_DB_API_KEY=...
JWT_SECRET=...
ENCRYPTION_KEY=...  # For platform API keys
FRONTEND_URL=http://localhost:3000
STRIPE_SECRET_KEY=...  # For subscription billing
STRIPE_WEBHOOK_SECRET=...
OPENAI_API_KEY=...  # Platform-managed (for embeddings)
AZURE_OPENAI_API_KEY=...  # Platform-managed (alternative)
AZURE_OPENAI_ENDPOINT=...
AZURE_OPENAI_DEPLOYMENT_ID=...
```

#### Workers (Python)
```env
KAFKA_BROKERS=localhost:9092
DATABASE_URL=postgresql://...
VECTOR_DB_TYPE=pinecone
VECTOR_DB_API_KEY=...
PLAYWRIGHT_BROWSER_PATH=/usr/bin/chromium
OPENAI_API_KEY=...  # Platform-managed (for CrewAI and embeddings)
CREWAI_LLM_MODEL=gpt-4  # For Advanced tier
```

---

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- [ ] Set up project structure (monorepo or separate repos)
- [ ] PostgreSQL schema and migrations
- [ ] Basic Golang API with authentication
- [ ] Next.js frontend setup with routing
- [ ] User registration and login

### Phase 2: Core Features (Weeks 3-4)
- [ ] Project CRUD operations
- [ ] Subscription tier system
- [ ] Platform API key management (encryption)
- [ ] Kafka setup and basic integration
- [ ] Index worker (Python) - consume jobs
- [ ] Vector DB integration (Pinecone/Weaviate)
- [ ] Basic indexing flow (end-to-end)
- [ ] Usage tracking infrastructure

### Phase 3: Real-time & Search (Weeks 5-6)
- [ ] WebSocket server (Golang)
- [ ] Real-time progress updates
- [ ] Search API and UI
- [ ] Log streaming
- [ ] Frontend state management

### Phase 4: MCP Integration & Billing (Weeks 7-8)
- [ ] MCP Agent implementation
- [ ] MCP server lifecycle management
- [ ] MCP status monitoring
- [ ] MCP server UI controls
- [ ] SSE execution tracking
- [ ] Usage limits enforcement
- [ ] Stripe integration for billing
- [ ] Subscription management UI

### Phase 5: Advanced Features (Weeks 9-10)
- [ ] CrewAI integration for Advanced tier
- [ ] Enhanced search_docs with AI
- [ ] find_code_snippet tool implementation
- [ ] Tier-based tool generation
- [ ] Usage analytics dashboard

### Phase 6: Polish & Production (Weeks 11-12)
- [ ] Error handling and retries
- [ ] Performance optimization
- [ ] Security hardening
- [ ] Documentation
- [ ] Deployment automation
- [ ] Monitoring and logging
- [ ] Billing webhook handling

---

## Scalability Considerations

### Horizontal Scaling

1. **API Servers**
   - Stateless design enables easy scaling
   - Load balancer distributes requests
   - Database connection pooling

2. **Workers**
   - Multiple index workers process jobs in parallel
   - Kafka partitions for job distribution
   - Auto-scaling based on queue depth

3. **MCP Servers**
   - Each server runs independently
   - Container orchestration for isolation
   - Resource limits per instance

### Performance Optimization

1. **Caching**
   - Redis for frequently accessed data
   - Cache project metadata
   - Cache search results (short TTL)

2. **Database**
   - Connection pooling
   - Read replicas for queries
   - Indexed queries only

3. **Vector Search**
   - Use hosted vector DB with built-in optimization
   - Batch embedding requests
   - Cache common queries

### Cost Optimization

1. **Vector DB**
   - Choose pricing model (per query vs. storage)
   - Archive old projects
   - Compress embeddings if supported

2. **Compute**
   - Spot instances for workers
   - Auto-scale down during low usage
   - Serverless for MCP (if applicable)

3. **Storage**
   - Clean up old job logs
   - Archive completed jobs
   - Compress stored data

---

## Additional Considerations

### Monitoring & Observability

- **Metrics**: Prometheus + Grafana
  - API request rates and latencies
  - Job processing times
  - MCP server health
  - Vector DB query performance

- **Logging**: Centralized logging (ELK, Loki)
  - Structured logging (JSON)
  - Log levels and filtering
  - Error tracking (Sentry)

- **Tracing**: Distributed tracing (Jaeger, Zipkin)
  - Request flow across services
  - Performance bottlenecks

### Testing Strategy

- **Unit Tests**: All services
- **Integration Tests**: API endpoints, database operations
- **E2E Tests**: Critical user flows (Playwright)
- **Load Tests**: Kafka, API, Vector DB

### Documentation

- **API Documentation**: OpenAPI/Swagger
- **Architecture Diagrams**: Updated with changes
- **Runbooks**: Deployment and troubleshooting
- **User Guide**: Frontend usage

---

## Billing Integration

### Stripe Integration

**Subscription Management**:
- Create Stripe products for each tier (Free, Basic, Pro, Advanced)
- Create Stripe prices (monthly/yearly)
- Handle subscription creation, updates, and cancellations
- Webhook handling for subscription events

**Webhook Events**:
- `customer.subscription.created` - New subscription
- `customer.subscription.updated` - Tier change
- `customer.subscription.deleted` - Cancellation
- `invoice.payment_succeeded` - Payment successful
- `invoice.payment_failed` - Payment failed

**Implementation**:
```go
// Golang webhook handler
func HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    event := stripe.Event{}
    // Verify webhook signature
    // Process event
    switch event.Type {
    case "customer.subscription.created":
        updateUserSubscription(event.Data.Object)
    case "customer.subscription.updated":
        updateUserSubscription(event.Data.Object)
    // ... other events
    }
}
```

**Usage-Based Billing** (Future):
- Track SSE executions per month
- Generate usage-based invoices
- Overage charges for exceeding limits
- Metered billing integration

### Payment Flow

1. **User selects tier** → Frontend
2. **Create Stripe Checkout Session** → Backend API
3. **User completes payment** → Stripe
4. **Webhook received** → Backend updates subscription
5. **User tier updated** → Database
6. **Features unlocked** → Immediate access

### Subscription States

- **Trial**: 14-day free trial (optional)
- **Active**: Paid subscription active
- **Cancelled**: Subscription cancelled, access until period end
- **Expired**: Subscription expired, downgrade to Free
- **Past Due**: Payment failed, grace period

---

## Open Questions & Decisions Needed

1. **Vector DB Choice**
   - Pinecone (easiest, managed)
   - Weaviate (self-hostable, more control)
   - Qdrant (open-source, self-hostable)

2. **MCP Server Deployment**
   - Subprocess (simple, less isolation)
   - Docker containers (better isolation, more overhead)
   - Serverless (cost-effective, cold starts)

3. **Multi-tenancy**
   - Single vector DB with namespace isolation
   - Separate collections per project
   - Separate vector DB instances per user (enterprise)

4. **Billing Provider**
   - Stripe (recommended, easiest integration)
   - Paddle (alternative)
   - Self-hosted billing (complex)

5. **SSE Execution Pricing**
   - Fixed monthly limits per tier
   - Overage charges (future)
   - Pay-as-you-go option (future)

6. **CrewAI Cost Management**
   - Cost per CrewAI execution
   - Rate limiting for Advanced tier
   - Caching AI responses

---

## Conclusion

This design document provides a comprehensive blueprint for converting the MCP-Docs CLI into a full-stack web application. The architecture is designed to be scalable, maintainable, and production-ready while preserving the core functionality of the existing CLI tool.

Key advantages of this design:
- **Subscription-Based SaaS**: No user API keys required, platform-managed
- **Tiered Access**: Clear feature differentiation across tiers
- **Usage-Based Billing**: SSE execution tracking for fair pricing
- **AI-Enhanced Features**: CrewAI integration for Advanced tier
- **Separation of Concerns**: Clear boundaries between layers
- **Scalability**: Horizontal scaling at every layer
- **Real-time Updates**: WebSocket for live progress
- **Flexibility**: Agent-based architecture for extensibility
- **Security**: Multi-layered security approach

Next steps:
1. Review and approve design
2. Set up development environment
3. Begin Phase 1 implementation
4. Iterate based on feedback

---

**Document Version**: 1.0  
**Last Updated**: 2024  
**Authors**: Development Team  
**Status**: Ready for Review

