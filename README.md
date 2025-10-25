# Cam Hub

A self-hosted smart camera hub for event-driven monitoring and analytics. Connect ONVIF/DVRIP IP cameras, detect motion, classify people via OpenVINO Model Server (OVMS), and view live streams and events in a React UI—fully containerized with Docker Compose.

## Key Features

- **Camera control & onboarding:** ONVIF PTZ and imaging controls; Wi-Fi pairing & onboarding via **DVRIP** (no vendor apps required).
- **Event pipeline:** GoCV motion detection + **FFmpeg** frame extraction → **OVMS** classification → **MinIO** object promotion (staging → detections/false-positives) → **PostgreSQL** metadata → positive detections published to clients via **Server-Sent Events (SSE)**.
- **Inter-process messaging:** **RabbitMQ** for decoupled communication between analyzer, motion, and API services.
- **Live viewing:** RTSP/WebRTC streaming via **MediaMTX**; React UI with real-time event feed and PTZ controls.
- **Deployment:** Single `docker compose up` stack (PostgreSQL, Redis, RabbitMQ, MediaMTX, MinIO, OVMS, Prometheus, Grafana).

## Architecture Overview

```text
Cameras (ONVIF/DVRIP)
   └─> motion (Go + GoCV) ──> analyzer (Go + FFmpeg + OVMS)
           │                           │
           └───────── RabbitMQ ────────┘
                                  │
                          PostgreSQL (metadata) + MinIO (objects)
                                  │
                   api (Go + Chi + SSE)  ──>  ui (React)  ──>  User
```

## Quickstart

1. **Clone**

```bash
git clone https://github.com/tomerab1/cam-hub && cd cam-hub
```

2. **Environment**

```bash
cp .env.example .env
# Fill in Postgres, MinIO, RabbitMQ, and OVMS settings
```

3. **Run**

```bash
docker compose up --build
```

4. **Access**

- API: `http://localhost:5555`
- UI: `http://localhost:5173` (or as configured)
- MediaMTX streams: as defined in compose configuration

## Services

- **api/** — REST + SSE endpoints for camera control and event streaming (Chi, Go).
- **analyzer/** — Handles FFmpeg frame extraction, OVMS inference, MinIO promotion, and RabbitMQ messaging.
- **motion/** — GoCV motion detection and frame triggering logic.
- **infra/** — Docker Compose setup, MediaMTX, OVMS configs.
- **ui/** — React app for live viewing, event log, PTZ controls.

## Planned Enhancements

- Prometheus exporters and Grafana dashboards (metrics and event visualization).
- Advanced UI filters, camera grouping, and event replay.
- Automated tests and profiling tools for multi-camera setups.

## Known Limitations

- Tested mainly with specific ONVIF/DVRIP camera models.
- OVMS configuration assumes `person-detection-retail-0013` model.

## License

MIT
