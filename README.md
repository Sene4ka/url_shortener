# url_shortener

Simple url-shortener api realisation, written in Golang

**Stack**:
- Go
- PgSQL

**Deploy**:
- Docker + Docker Compose

## **Starting shortener API tutorial**

```bash
# 1. Clone + setup
git clone https://github.com/Sene4ka/url_shortener.git
cd url_shortener

# 2. Environment
cp .env.example .env
# Set your desired env values for docker compose

# 3. Start the server
make docker-up 
# Automatically Executes 'make build' and 'make build-docker' before starting API with 'docker compose'
# Also checks for DB_USE_IN_MEMORY(true/false, default true) .env var
# If true starts compose without Postgres containers, otherwise, vice versa

# 4. For more options
make help