Livetran
========

Self‑hosted live streaming server in Go. Ingest via SRT, transcode with FFmpeg to HLS, serve locally, and upload to Cloudflare R2 for scalable delivery. Secure APIs with HMAC request signing and JWT stream keys. Optional OpenTelemetry metrics.

Contents
--------
- Overview
- Features
- Architecture
- Quick start
- Configuration (.env)
- Running (Docker and local)
- API reference (start/stop/status)
- Security (HMAC + JWT stream keys)
- Video playback
- Webhooks
- Metrics and observability
- Deployment notes
- License

Overview
--------
Livetran exposes a simple HTTP API to manage a live stream lifecycle:
- Start a stream: creates an SRT listener on a random free port and generates a JWT‑backed stream key.
- Ingest: your encoder (e.g., OBS) publishes to the returned SRT URL.
- Transcode: FFmpeg converts the incoming SRT MPEG‑TS to HLS segments and playlists under `output/`.
- Upload: a watcher pushes `.ts` and `.m3u8` files to Cloudflare R2; the first public playlist URL is returned via webhook.
- Serve: HLS files are available locally under `/video/` for testing, or via your R2 public URL in production.

Features
--------
- Secure SRT ingestion with JWT stream keys
- Simple REST API for start/stop/status
- FFmpeg HLS transcoding (single‑profile or ABR ladder)
- Cloudflare R2 uploads (S3‑compatible)
- Real‑time webhooks on status updates
- CORS enabled, HMAC‑SHA256 request verification
- Optional OpenTelemetry metrics export

Architecture
------------
See docs guide for details and diagram:
- `docs/guide/introduction.mdx`
- `docs/arch.svg`

Quick start
-----------
Prerequisites:
- Go 1.21+ if running locally
- FFmpeg installed (Docker image includes it)
- Cloudflare R2 bucket and credentials
- TLS keypair at `keys/localhost.pem` and `keys/localhost-key.pem` (self‑signed is fine for dev)

Clone and prepare `.env`:
```bash
cp .env.example .env   # if you keep a template; otherwise create .env with the vars below
```

Configuration (.env)
--------------------
Required:
- JWT_SECRET: HMAC secret for stream key JWT
- HMAC_SECRET: HMAC secret for signing REST request bodies
- R2_ACCOUNT_ID: Cloudflare account id for R2 (used in endpoint)
- R2_ACCESS_KEY: Cloudflare R2 access key id
- R2_SECRET_KEY: Cloudflare R2 secret access key
- BUCKET_NAME: Cloudflare R2 bucket to upload HLS artifacts
- CLOUDFLARE_PUBLIC_URL: Base public URL that serves your R2 objects (e.g., https://r2.example.com/hls)

Optional (metrics):
- ENABLE_METRICS: set `true` to enable OTLP metrics export
- OTEL_EXPORTER_OTLP_ENDPOINT: default `localhost:4318`
- OTEL_EXPORTER_OTLP_INSECURE: `true` to disable TLS for exporter (default `true`)
- SERVICE_VERSION, ENV: resource attributes for metrics

Running (Docker)
----------------
Build and run:
```bash
docker build -t livetran .
docker run -d \
  -p 8080:8080 \
  --name livetran \
  --env-file .env \
  -v "$(pwd)/output:/app/output" \
  -v "$(pwd)/keys:/app/keys:ro" \
  livetran
```

Running (local)
---------------
```bash
go mod download
go run ./cmd/main.go
```

The server listens on HTTPS at `:8080` and expects TLS keys at `keys/localhost.pem` and `keys/localhost-key.pem`.

API reference
-------------
Base path: `/api` (all endpoints require HMAC request signing; see Security)

1) Start stream
```http
POST /api/start-stream
Content-Type: application/json
LT-SIGNATURE: <hex(hmac_sha256(body,HMAC_SECRET))>

{"stream_id":"req1","webhook_urls":["https://example.com/webhook"],"abr":true}
```
Response:
```json
{"success":true,"data":"Stream launching!"}
```

2) Stop stream
```http
POST /api/stop-stream
Content-Type: application/json
LT-SIGNATURE: <hex(hmac_sha256(body,HMAC_SECRET))>

{"stream_id":"req1"}
```
Response:
```json
{"success":true,"data":"Stream stopped!"}
```

3) Status
```http
GET /api/status
Content-Type: application/json
LT-SIGNATURE: <hex(hmac_sha256(body,HMAC_SECRET))>

{"stream_id":"req1"}
```
Response (example):
```json
{"success":true,"data":"Status: STREAMING"}
```

Security
--------
HMAC request signing (all `/api/*` routes):
- Compute `hex(hmac_sha256(<raw body>, HMAC_SECRET))` and set header `LT-SIGNATURE`.
- Requests without a valid signature are rejected.

JWT stream keys (SRT publish):
- When you start a stream, Livetran generates a JWT stream key for the given `stream_id` using `JWT_SECRET`.
- Your encoder connects using the returned URL template:
  `srt://<server_ip>:<port>?streamid=mode=publish,rid=<stream_id>,token=<jwt>`
- The server validates that `token` is valid, unexpired, and matches `rid`.

Video playback
--------------
Local testing endpoint (serves files from `output/`):
- HLS playlists/chunks: `GET /video/<file>`
  - Content types: `.m3u8` => `application/vnd.apple.mpegurl`, `.ts` => `video/MP2T`
In production, serve HLS from your Cloudflare R2 public URL.

ABR vs single‑profile
---------------------
- Set `abr=true` in the start request to enable an HLS variant ladder (1080p/720p/480p), with a master playlist named `<stream_id>_master.m3u8`.
- If `abr=false` (default), a single playlist `<stream_id>.m3u8` is produced.

Webhooks
--------
Provide one or more `webhook_urls` in `start-stream` to receive JSON updates. Example payload:
```json
{
  "Status": "READY|STREAMING|STOPPED",
  "Update": "message",
  "StreamLink": "https://.../req1_master.m3u8"
}
```
Notes:
- The first time a public playlist is uploaded, `StreamLink` is included.
- On ABR, link is emitted when the master playlist is available.

Metrics and observability
-------------------------
- Set `ENABLE_METRICS=true` to enable OpenTelemetry metrics export over OTLP/HTTP.
- Configure exporter via `OTEL_EXPORTER_OTLP_ENDPOINT` and `OTEL_EXPORTER_OTLP_INSECURE`.
- A gauge `streams_info{status=idle|active|stopped}` reports counts derived from the in‑memory `TaskManager`.
- Sample Grafana/Prometheus/Loki/OTel Collector configs are under `metrics/deployment/`.

Deployment notes
----------------
- Ensure valid TLS certs in `keys/` for HTTPS server startup.
- Persist `output/` if you want local playback beyond container lifecycle (Docker volume provided).
- `.gitignore` should exclude `output/`, secrets, and local artifacts; keep `keys/` secure.

License
-------
Apache 2.0 — see `LICENSE`.

