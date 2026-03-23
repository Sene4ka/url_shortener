# url_shortener

Simple url-shortener api realisation, written in Golang

**Stack**:
- Go
- PgSQL

**Deploy**:
- Docker, Docker Compose, Docker Hub

## **Starting shortener API tutorial**

### Pulling from Docker Hub

```bash
# 1. Pull image from Docker Hub
docker pull s3ne4ka/shortener:latest

# 2. Start container changing env variables and exposed ports if needed
# Full env variables list with their default values you can see in .env.example
# Example of changing env var and exposed ports:
docker run -d -p 8080:8080 -e DB_USE_IN_MEMORY=true s3ne4ka/shortener:latest 
```

### Manual Build and Run

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
```
