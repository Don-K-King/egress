# Local Egress deployment

This compose setup assumes LiveKit is already running and exposes its services on a shared external Docker network.

## Prerequisites

1. Create the shared network once:

   ```bash
   docker network create livekit-net
   ```

2. Start the LiveKit stack attached to `livekit-net` (including `livekit` and `redis` services).

3. Create a local Egress config from the example and fill in API credentials that match LiveKit:

   ```bash
   cp deploy/local/egress.config.yaml.example deploy/local/egress.config.yaml
   ```

   Required values:

   - `ws_url: ws://livekit:7880`
   - `redis.address: redis:6379`
   - `api_key`/`api_secret` must match the LiveKit stack

## Start Egress

```bash
docker compose -f deploy/local/docker-compose.yml up -d
```

Egress will join `livekit-net` and reuse Redis from the LiveKit stack.
