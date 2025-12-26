# Environment Variables Setup

This project uses environment variables to keep sensitive information out of the codebase. Follow these steps to set up your environment:

## Quick Setup

1. **Copy the example file:**
   ```bash
   cp .env.example .env
   ```

   On Windows PowerShell:
   ```powershell
   Copy-Item .env.example .env
   ```

2. **Edit the `.env` file** and replace the placeholder values with your actual secrets:
   - `POSTGRES_PASSWORD`: Change `your_secure_password_here` to a strong password
   - Update any other values as needed for your environment

3. **Verify `.env` is in `.gitignore`** (it should already be there)

## Environment Variables

The following environment variables are used in `docker-compose.yml`:

### PostgreSQL
- `POSTGRES_USER`: Database username (default: `mcpdocs`)
- `POSTGRES_PASSWORD`: **CHANGE THIS** - Database password
- `POSTGRES_DB`: Database name (default: `mcp_docs`)
- `POSTGRES_HOST`: Database host (default: `postgres` for Docker)
- `POSTGRES_PORT`: Database port (default: `5432`)

### Go Backend
- `GO_BACKEND_PORT`: Backend API port (default: `8080`)

### Frontend
- `NEXT_PUBLIC_API_URL`: API endpoint URL (default: `http://localhost:8080`)
- `FRONTEND_PORT`: Frontend port (default: `3000`)

### Kafka (Optional)
- `KAFKA_BROKER_ID`: Kafka broker ID (default: `1`)
- `KAFKA_ADVERTISED_LISTENERS`: Kafka listeners configuration

### Observability (Optional)
- `METRICS_PORT`: Python RAG Agent Prometheus metrics port (default: `9091`)
- Go backend exposes metrics on port `9090`
- Prometheus UI is accessible on port `9090` (host)
- Grafana UI is accessible on port `3001` (host) with credentials `admin/admin`

## Security Notes

- ✅ `.env` is already in `.gitignore` - it will NOT be committed to Git
- ✅ `.env.example` is safe to commit - it contains no real secrets
- ⚠️ **Never commit your `.env` file to version control**
- ⚠️ **Change all default passwords before deploying to production**

## Docker Compose

The `docker-compose.yml` file automatically loads variables from `.env` using the `${VARIABLE_NAME}` syntax. If a variable is not set, it will use the default value specified with `:-` (e.g., `${PORT:-8080}`).

