# Production Deployment Checklist — AniQuiz

## Pre-Deploy

### Secrets & Configuration
- [ ] `POSTGRES_PASSWORD` set (min 32 chars, random)
- [ ] `JWT_SECRET` set (min 64 chars, random — `openssl rand -hex 64`)
- [ ] `FRONTEND_URL` matches the exact production origin (e.g. `https://aniquiz.example.com`)
- [ ] OAuth credentials configured (AniList + MAL redirect URIs updated to prod domain)
- [ ] `GRAFANA_PASSWORD` set
- [ ] All secrets stored in GitHub Environments (not in `.env` committed to repo)
- [ ] `.env` file on server has `chmod 600`

### Infrastructure
- [ ] Server has at least 2 GB RAM, 20 GB disk
- [ ] Docker and Docker Compose v2 installed
- [ ] Firewall: only ports 80, 443, 22 open externally
- [ ] Ports 9090, 3000, 8081, 9187 bound to `127.0.0.1` only (monitoring internal)
- [ ] SSH key-based auth only (password auth disabled)
- [ ] `fail2ban` or equivalent installed

### TLS / Domain
- [ ] DNS A record pointing to server IP
- [ ] Run initial Let's Encrypt certificate:
  ```bash
  docker run --rm -p 80:80 \
    -v letsencrypt_certs:/etc/letsencrypt \
    -v certbot_webroot:/var/www/certbot \
    certbot/certbot certonly --standalone -d aniquiz.example.com --email you@example.com --agree-tos
  ```
- [ ] Update `DOMAIN` in `nginx/conf.d/aniquiz.conf`
- [ ] Verify HTTPS works: `curl -I https://aniquiz.example.com`
- [ ] Verify HSTS header is present

### Database
- [ ] PostgreSQL data volume on persistent storage (not ephemeral)
- [ ] First-run migration completes without error
- [ ] Backup script scheduled: `crontab -e` → `0 3 * * * /opt/aniquiz/scripts/backup-db.sh`
- [ ] Test restore from backup before first deploy
- [ ] Backups stored off-server (rsync/S3/Backblaze)

---

## Deploy

```bash
cd /opt/aniquiz

# Start app + Nginx
IMAGE_TAG=sha-<commit> docker compose -f docker-compose.prod.yml up -d

# Start monitoring stack
docker compose -f docker-compose.monitoring.yml up -d

# Verify health
curl https://aniquiz.example.com/health
```

Or use the deploy script for zero-downtime + rollback:
```bash
./scripts/deploy.sh sha-<commit>
```

---

## Post-Deploy Verification

### Functional
- [ ] `GET /health` returns `{"status":"ok"}`
- [ ] User registration works
- [ ] User login returns JWT
- [ ] Room creation works
- [ ] WebSocket connects (`/ws?room=...&token=...`)
- [ ] Leaderboard loads
- [ ] OAuth redirect URLs correct (AniList + MAL)
- [ ] Static assets load (favicon, JS bundles)
- [ ] SPA routing works (deep link → index.html)

### Security
- [ ] `curl -I http://aniquiz.example.com` → 301 redirect (not 200)
- [ ] HSTS header present on HTTPS response
- [ ] `X-Frame-Options: DENY` present
- [ ] `/test-audio` returns 404 (debug route disabled in production)
- [ ] Rate limiting works: `for i in $(seq 1 20); do curl -s -o /dev/null -w "%{http_code}\n" -X POST https://aniquiz.example.com/api/auth/login; done`
- [ ] WebSocket from wrong origin is rejected (401)

### Monitoring
- [ ] Grafana accessible at `http://server:3000` (internal only)
- [ ] Prometheus targets all UP: `http://server:9090/targets`
- [ ] cAdvisor shows container metrics
- [ ] Loki receiving logs: check Grafana → Explore → Loki

### CI/CD
- [ ] Push to `main` triggers full pipeline
- [ ] Docker image pushed to GHCR with `sha-` + `latest` tags
- [ ] Trivy SARIF uploaded to GitHub Security tab
- [ ] GitHub Environment `staging` and `production` configured with required reviewers
- [ ] Deploy SSH key added to GitHub Secrets (`DEPLOY_SSH_KEY`, `STAGING_HOST` / `PROD_HOST`, `DEPLOY_USER`)

---

## Known Scaling Limitation

The WebSocket game rooms are stored in-memory (`internal/game/room.go`). This means:

- **Single-replica only** until Redis pub/sub is integrated
- Do **not** increase `replicas > 1` in `k8s/app/deployment.yml` or the HPA `minReplicas`
- Players on different pods will be in isolated rooms

**To enable horizontal scaling:**
1. Add Redis to the stack
2. Replace `game.ActiveRooms` (map) with a Redis-backed room registry
3. Replace in-process broadcast with Redis pub/sub per room
4. Then set `minReplicas: 2` in the HPA

---

## Rollback

```bash
# Manual rollback to previous image
./scripts/deploy.sh sha-<previous-commit>

# Or Kubernetes
kubectl set image deployment/aniquiz aniquiz=ghcr.io/leyapson/aniquiz:<previous-tag> -n aniquiz
kubectl rollout status deployment/aniquiz -n aniquiz
```

## Useful Commands

```bash
# View live app logs
docker compose -f docker-compose.prod.yml logs -f app

# View Nginx access logs
docker compose -f docker-compose.prod.yml logs -f nginx

# DB shell
docker exec -it aniquiz-db-1 psql -U aniquiz

# Force Nginx config reload (no downtime)
docker exec aniquiz-nginx-1 nginx -s reload

# Check certificate expiry
docker exec aniquiz-nginx-1 openssl x509 -enddate -noout \
  -in /etc/letsencrypt/live/aniquiz.example.com/cert.pem
```
