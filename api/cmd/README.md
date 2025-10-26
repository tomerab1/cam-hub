# Cam-Hub API Services

This directory contains the core executable services for the **Cam-Hub** system - a self-hosted smart camera hub for event-driven monitoring and analytics. The system connects ONVIF/DVRIP IP cameras, detects motion, classifies people via OpenVINO Model Server (OVMS), and provides live streaming with a React UI.

## System Overview

Cam-Hub implements a sophisticated event-driven architecture with four main services:

1. **API Server** - REST API + Server-Sent Events for camera control and real-time streaming
2. **Motion Detection** - Real-time RTSP monitoring with GoCV-based motion detection
3. **Frame Analyzer** - AI-powered frame analysis using OpenVINO Model Server
4. **Supervisor** - Process lifecycle management for motion detection instances

## Services

### 1. API Server (`api/main.go`)

The central API service that orchestrates the entire camera ecosystem.

**Core Functionality:**
- **Camera Discovery & Pairing**: ONVIF-based camera discovery with automatic pairing
- **WiFi Configuration**: DVRIP protocol integration for camera WiFi setup
- **PTZ Control**: Pan-Tilt-Zoom operations via ONVIF
- **Stream Management**: MediaMTX integration for RTSP/WebRTC streaming
- **Real-time Events**: Server-Sent Events (SSE) for live camera status and alerts
- **User Management**: Camera credential management and authentication

**API Endpoints:**
```
GET    /cameras/discovery           # Discover unpaired cameras
GET    /cameras                     # List paired cameras
POST   /cameras/{uuid}/pair         # Pair a camera
DELETE /cameras/{uuid}/pair         # Unpair a camera
GET    /cameras/{uuid}/stream       # Get camera stream URL
DELETE /cameras/{uuid}/stream       # Delete camera stream
POST   /cameras/{uuid}/ptz/move     # Move camera (PTZ)
GET    /events/discovery            # SSE: Camera discovery events
GET    /events/recordings/{uuid}    # SSE: Motion detection alerts
```

**Key Features:**
- **ONVIF Integration**: Full ONVIF protocol support for camera control
- **DVRIP Support**: WiFi pairing without vendor apps
- **MediaMTX Integration**: Automatic stream publishing and management
- **Real-time SSE**: Live camera discovery and motion alerts
- **Scheduled Discovery**: Automatic camera discovery every 30 seconds
- **Credential Management**: Secure storage of camera credentials
- **Process Coordination**: RabbitMQ-based communication with other services

**Dependencies:**
- PostgreSQL (camera metadata, credentials)
- Redis (camera status cache)
- RabbitMQ (inter-service messaging)
- MediaMTX (streaming server)

**Environment Variables:**
```bash
SERVER_ADDR=:5555                    # HTTP server address
POSTGRES_DSN=postgres://...          # Database connection
REDIS_CACHE=localhost:6379           # Redis server
REDIS_PASSWORD=                       # Redis password
RABBITMQ_ADDR=amqp://...             # RabbitMQ connection
CAMERA_GLOB_ADMIN_USERNAME=admin     # Default camera admin user
CAMERA_GLOB_ADMIN_PASS=admin         # Default camera admin password
LOGGER_PATH=/var/log/cam-hub         # Log directory
```

### 2. Motion Detection (`motion_detection/main.go`)

Real-time motion detection service using computer vision algorithms.

**Core Functionality:**
- **RTSP Stream Processing**: Continuous monitoring of camera streams
- **Motion Detection**: GoCV-based background subtraction with MOG2 algorithm
- **Frame Capture**: Automatic frame extraction during motion events
- **Video Processing**: FFmpeg-based video concatenation and frame extraction
- **Object Storage**: MinIO integration for frame and video storage
- **Analysis Pipeline**: RabbitMQ integration for frame analysis

**Motion Detection Algorithm:**
```go
// Background subtraction with MOG2
mog2 := gocv.NewBackgroundSubtractorMOG2WithParams(1024, 16, false)
// Threshold-based motion detection
threshold := 50.0  // Configurable sensitivity
minAreaPixels := 15000  // Minimum motion area
warmupFrames := 150     // Background stabilization
```

**Processing Pipeline:**
1. **Motion Detection**: Continuous RTSP stream analysis
2. **Event Triggering**: Motion detected with cooldown period (10s)
3. **Video Assembly**: Concatenate 3 video segments around motion event
4. **Frame Extraction**: Extract 4 frames at 1fps, scaled to 544x320
5. **Storage**: Upload to MinIO staging area
6. **Analysis Queue**: Send to frame analyzer via RabbitMQ

**Key Features:**
- **Real-time Processing**: 50ms frame analysis interval
- **Motion Cooldown**: Prevents spam with 10-second cooldown
- **FFmpeg Integration**: Video concatenation and frame extraction
- **Scalable Processing**: Concurrent motion detection per camera
- **MinIO Storage**: Automatic upload to object storage
- **RabbitMQ Integration**: Event-driven analysis pipeline

**Usage:**
```bash
go run cmd/motion_detection/main.go -addr rtsp://camera-ip:554/stream
```

**Dependencies:**
- OpenCV (GoCV) for computer vision
- FFmpeg for video processing
- MinIO for object storage
- RabbitMQ for messaging

**Environment Variables:**
```bash
MINIO_BUCKET_NAME=cam-hub-storage     # MinIO bucket name
RABBITMQ_ADDR=amqp://...              # RabbitMQ connection
LOGGER_PATH=/var/log/cam-hub          # Log directory
```

### 3. Frame Analyzer (`frame_analyzer/main.go`)

AI-powered frame analysis service using OpenVINO Model Server.

**Core Functionality:**
- **AI Inference**: OpenVINO Model Server integration for person detection
- **Tensor Processing**: Batch processing of 4 frames per motion event
- **Object Detection**: Person detection with confidence scoring
- **Storage Management**: MinIO object promotion based on detection results
- **Metadata Management**: PostgreSQL integration for recording metadata
- **Event Broadcasting**: Real-time alerts for positive detections

**AI Processing Pipeline:**
1. **Frame Loading**: Load 4 frames from MinIO staging area
2. **Tensor Construction**: Convert frames to 544x320x3 float32 tensors
3. **Batch Processing**: Process 4 frames in single inference call
4. **Model Inference**: `person-detection-retail-0013` model via gRPC
5. **Result Analysis**: Extract confidence scores and bounding boxes
6. **Storage Decision**: Promote to detections or false-positives based on confidence
7. **Event Broadcasting**: Send alerts for detections ≥0.5 confidence

**Model Specifications:**
```go
const (
    tensorW = 544    // Input width
    tensorH = 320    // Input height  
    tensorC = 3      // RGB channels
    tensorN = 4      // Batch size
)
```

**Storage Strategy:**
- **Staging Area**: Initial frame storage
- **Detections**: High-confidence detections (≥0.5) → long retention
- **False Positives**: Low-confidence detections (<0.5) → short retention
- **Metadata**: PostgreSQL for searchable recording information

**Key Features:**
- **Batch Processing**: Efficient 4-frame batch inference
- **Confidence Thresholding**: 0.5 confidence threshold for positive detections
- **Automatic Promotion**: MinIO object lifecycle management
- **Real-time Alerts**: SSE broadcasting for positive detections
- **Metadata Storage**: Comprehensive recording information
- **Retention Management**: Configurable retention periods

**Dependencies:**
- OpenVINO Model Server (OVMS)
- PostgreSQL database
- MinIO object storage
- RabbitMQ message broker

**Environment Variables:**
```bash
OVMS_GRPC_ADDR=localhost:9000        # OpenVINO Model Server gRPC
POSTGRES_DSN=postgres://...          # Database connection
MINIO_BUCKET_NAME=cam-hub-storage     # MinIO bucket
MINIO_STAGING_KEY=staging             # Staging area prefix
MINIO_DETECTIONS_KEY=detections       # Positive detections prefix
MINIO_FALSE_POSITIVES_KEY=false-positives  # False positives prefix
MINIO_DETECTIONS_DAYS=30             # Detection retention days
MINIO_FALSE_POSITIVES_DAYS=7         # False positive retention days
RABBITMQ_ADDR=amqp://...             # RabbitMQ connection
LOGGER_PATH=/var/log/cam-hub         # Log directory
```

### 4. Supervisor (`supervisor/main.go`)

Process lifecycle management service for motion detection instances.

**Core Functionality:**
- **Process Management**: Spawn and manage motion detection processes per camera
- **Lifecycle Control**: Automatic process registration and cleanup
- **Event Handling**: Camera pairing/unpairing event processing
- **Resource Management**: Process monitoring and cleanup
- **Graceful Shutdown**: Proper process termination on system shutdown

**Process Management:**
```go
// Process registration with camera-specific arguments
supervisor.NotifyCtrl(visor.CtrlEvent{
    Kind:    visor.CtrlRegister,
    CamUUID: cameraUUID,
    Args:    []string{"run", "./cmd/motion_detection", "-addr", streamURL},
})
```

**Event Processing:**
1. **Camera Paired**: Spawn new motion detection process
2. **Camera Unpaired**: Terminate motion detection process
3. **Process Exit**: Handle process failures and cleanup
4. **Version Management**: Track camera revisions for updates

**Key Features:**
- **Dynamic Process Spawning**: One process per camera
- **Event-driven Management**: RabbitMQ-based camera lifecycle events
- **Process Monitoring**: Automatic cleanup on process exit
- **Graceful Termination**: SIGTERM/SIGKILL handling
- **Resource Cleanup**: Proper process group management
- **Error Handling**: Robust process failure recovery

**Process Lifecycle:**
1. **Registration**: Camera paired → spawn motion detection process
2. **Monitoring**: Continuous process health monitoring
3. **Cleanup**: Camera unpaired → terminate process gracefully
4. **Recovery**: Process exit → cleanup and optional restart

**Dependencies:**
- RabbitMQ message broker

**Environment Variables:**
```bash
RABBITMQ_ADDR=amqp://...             # RabbitMQ connection
RABBITMQ_PAIR_KEY=camera.paired      # Camera pairing queue
RABBITMQ_UNPAIR_KEY=camera.unpaired  # Camera unpairing queue
LOGGER_PATH=/var/log/cam-hub         # Log directory
```

## System Architecture

Cam-Hub implements a sophisticated event-driven microservices architecture:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cam-Hub System                           │
├─────────────────────────────────────────────────────────────────┤
│  React UI (Port 5173)                                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Live Streams │ Event Feed │ PTZ Controls │ Camera Mgmt  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼ HTTP/SSE
┌─────────────────────────────────────────────────────────────────┐
│                    API Server (Port 5555)                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ ONVIF Discovery │ Camera Pairing │ PTZ Control │ SSE    │   │
│  │ DVRIP WiFi     │ Stream Mgmt    │ Credentials │ Events │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   PostgreSQL    │  │     Redis       │  │    RabbitMQ     │
│                 │  │                 │  │                 │
│ Camera Metadata │  │ Status Cache    │  │ Event Bus       │
│ Credentials     │  │ Discovery Data  │  │ Message Queue   │
│ Recordings      │  │ Session Data    │  │ Coordination    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
                                │
                                ▼ Events
┌─────────────────────────────────────────────────────────────────┐
│                    Supervisor Service                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Process Lifecycle │ Camera Events │ Resource Mgmt      │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼ Spawn Processes
┌─────────────────────────────────────────────────────────────────┐
│              Motion Detection Instances (Per Camera)           │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ RTSP Monitor │ GoCV Motion │ FFmpeg Process │ MinIO    │   │
│  │ Background   │ Detection   │ Video/Frames   │ Upload   │   │
│  │ Subtraction  │ Threshold   │ Extraction     │ Storage  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼ RabbitMQ Messages
┌─────────────────────────────────────────────────────────────────┐
│                    Frame Analyzer Service                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ OpenVINO Model │ Tensor Process │ AI Inference │       │   │
│  │ Server (OVMS)  │ Batch Frames   │ Person Detect │      │   │
│  │ gRPC Client    │ 544x320x3      │ Confidence    │      │   │
│  │ Model: retail  │ Float32 Tensors│ Thresholding  │      │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│     MinIO       │  │   PostgreSQL    │  │    RabbitMQ     │
│                 │  │                 │  │                 │
│ Object Storage  │  │ Recording Meta  │  │ Detection Events│
│ Staging Area    │  │ Evidence Data   │  │ SSE Broadcasting│
│ Detections      │  │ Retention Info  │  │ Real-time Alerts│
│ False Positives │  │ Search Index    │  │ Client Notify   │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## Event Flow & Data Pipeline

### 1. Camera Discovery & Pairing
```
ONVIF Discovery → API Server → Database → Supervisor → Motion Detection Process
```

### 2. Motion Detection Pipeline
```
RTSP Stream → GoCV Motion Detection → FFmpeg Processing → MinIO Storage → RabbitMQ → Frame Analyzer
```

### 3. AI Analysis Pipeline
```
MinIO Frames → Tensor Construction → OVMS Inference → Confidence Analysis → Storage Promotion → SSE Alerts
```

### 4. Real-time Event Broadcasting
```
Motion Detection → RabbitMQ → API Server → SSE → React UI
```

## Technology Stack

### Core Technologies
- **Go 1.24+**: Primary language for all services
- **OpenCV (GoCV)**: Computer vision and motion detection
- **FFmpeg**: Video processing and frame extraction
- **OpenVINO Model Server**: AI inference for person detection
- **PostgreSQL**: Relational database for metadata
- **Redis**: Caching and session management
- **RabbitMQ**: Message broker for inter-service communication
- **MinIO**: Object storage for videos and frames
- **MediaMTX**: RTSP/WebRTC streaming server

### Protocols & Standards
- **ONVIF**: Camera discovery and control protocol
- **DVRIP**: Camera WiFi configuration protocol
- **RTSP**: Real-time streaming protocol
- **WebRTC**: Browser-based streaming
- **Server-Sent Events (SSE)**: Real-time event streaming
- **gRPC**: AI model inference communication

## Building and Running

### Prerequisites

**System Requirements:**
- Go 1.24.0+
- Docker & Docker Compose
- FFmpeg (for video processing)
- OpenCV libraries (for motion detection)

**External Services:**
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+
- MinIO (latest)
- OpenVINO Model Server
- MediaMTX

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/tomerab1/cam-hub && cd cam-hub

# Configure environment
cp .env.example .env
# Edit .env with your database and service configurations

# Start all services
docker compose up --build
```

### Manual Build & Run

```bash
# Build all services
go build -o bin/api cmd/api/main.go
go build -o bin/motion_detection cmd/motion_detection/main.go
go build -o bin/frame_analyzer cmd/frame_analyzer/main.go
go build -o bin/supervisor cmd/supervisor/main.go

# Start services in order
./bin/api &
./bin/supervisor &
./bin/frame_analyzer &

# Motion detection instances are automatically spawned by supervisor
```

### Service Startup Order

1. **Infrastructure Services**: PostgreSQL, Redis, RabbitMQ, MinIO, OVMS, MediaMTX
2. **API Server**: Main HTTP API and camera management
3. **Supervisor**: Process lifecycle management
4. **Frame Analyzer**: AI processing service
5. **Motion Detection**: Automatically spawned per camera by supervisor

## Configuration

### Environment Variables

Each service requires specific environment variables. See individual service sections above for complete configuration details.

### Key Configuration Files

- **`.env`**: Main environment configuration
- **`compose.yml`**: Docker Compose service definitions
- **`mediamtx.yml`**: MediaMTX streaming configuration
- **`models/config.json`**: OpenVINO Model Server configuration

## Monitoring & Observability

### Logging
- **Structured JSON Logs**: All services use structured logging
- **Log Rotation**: Automatic rotation at 25MB with 14-day retention
- **Compression**: Logs are compressed after rotation
- **Service-specific Logs**: Each service logs to separate files

### Health Monitoring
- **API Health Endpoints**: HTTP health checks available
- **Process Monitoring**: Supervisor tracks process health
- **Real-time Events**: SSE streams for live monitoring
- **Database Health**: Connection monitoring and retry logic

### Metrics & Alerting
- **Motion Detection Metrics**: Frame processing rates and detection counts
- **AI Inference Metrics**: Model performance and confidence distributions
- **Storage Metrics**: MinIO usage and retention management
- **System Metrics**: Process health and resource utilization

## Development & Debugging

### Development Setup
1. **Local Development**: Use Docker Compose for external dependencies
2. **Service Debugging**: Run individual services with debug logging
3. **API Testing**: Use the `/test` endpoint for DVRIP client testing
4. **Motion Testing**: Use test cameras or RTSP streams for motion detection

### Common Debugging Steps
1. **Check Logs**: Review service-specific log files for errors
2. **Verify Dependencies**: Ensure all external services are running
3. **Test Connectivity**: Verify camera RTSP streams and network connectivity
4. **Monitor Queues**: Check RabbitMQ management UI for message flow
5. **Storage Verification**: Confirm MinIO bucket access and permissions

### Performance Tuning
- **Motion Detection**: Adjust threshold and cooldown parameters
- **AI Processing**: Configure batch sizes and confidence thresholds
- **Storage**: Optimize retention policies and compression settings
- **Streaming**: Tune MediaMTX configuration for optimal performance

## Security Considerations

- **Camera Credentials**: Stored securely in PostgreSQL with encryption
- **Network Security**: Use VPN or secure networks for camera access
- **API Security**: Implement authentication and rate limiting
- **Storage Security**: Configure MinIO access policies and encryption
- **Process Isolation**: Motion detection processes run in isolated environments
