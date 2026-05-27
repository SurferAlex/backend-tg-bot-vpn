# Deploy notes

## Monitor + Redis (docker-compose)

1. Copy `monitor/.env.example` → `monitor/.env` (tokens aligned with VpnAPI and bot).
2. Set host paths:
   - `XRAY_ACCESS_LOG_HOST_DIR` — directory with Xray `access.log` (e.g. `/opt/3x-ui/logs` if mounted as `access.log`).
3. `docker compose up -d redis monitor vpnapi bot`

Alert dedup key: `alerted:{client_uuid}:{ip}` (TTL `MONITOR_ALERT_DEDUP_TTL`, default 30m).

## Fail2ban on VPS (host)

For **3x-ui** IP limit: use the jail/integration that 3x-ui documents (logs `[LIMIT_IP] ... queued for fail2ban`).
