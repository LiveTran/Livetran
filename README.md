
# LiveTran 📹

**LiveTran** is a self-hostable, high-performance live streaming media server written in Go. It's designed to ingest video streams via SRT (Secure Reliable Transport), transcode them in real-time to HLS (HTTP Live Streaming), and deliver them to viewers with low latency.

---

## ✨ Features

- **🎬 SRT & WebRTC Ingest:** Secure, low-latency video ingest using industry-standard protocols.
- **⚡ Real-time Transcoding:** On-the-fly transcoding to multi-bitrate HLS using FFmpeg for adaptive streaming.
- **☁️ Cloud-Native:** Seamless integration with Cloudflare R2 for scalable, cost-effective segment storage and delivery.
- **🔐 Secure by Design:** Protect your streams with JWT-based authentication and signed playback URLs.
- **📈 Scalable Architecture:** Built on a modular, container-friendly architecture ready for Kubernetes deployment.
- **📊 Real-time Monitoring:** Prometheus and Grafana integration for at-a-glance stream health and performance analytics.

## 🏛️ Architecture

LiveTran follows a microservice-based architecture that separates ingest, transcoding, and delivery into distinct, scalable components.

```
Streaming Client (OBS)
       |
       | SRT/WebRTC
       v
+---------------------+
|   LiveTran Server   |
| +-----------------+ |
| | Ingest Service  | |
| +-----------------+ |
|         |           |
|         v           |
| +-----------------+ |
| |Transcoding Engine| |
| +-----------------+ |
|         |           |
|         v           |
|  (HLS Segments)     |
+---------------------+
       |
       v
+--------------------------+
| Cloudflare R2 Storage    |
+--------------------------+
       |
       | Signed URLs
       v
+--------------------------+
|   CDN (Cloudflare)       |
+--------------------------+
       |
       | HLS Playback
       v
    Viewer
```

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (v1.21+)
- [FFmpeg](https://ffmpeg.org/download.html)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)

### Quickstart with Docker

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/your-username/LiveTran.git
    cd LiveTran
    ```

2.  **Configure Environment:**
    Copy the example environment file and update it with your Cloudflare R2 credentials and other settings.
    ```sh
    cp .env.example .env
    ```

3.  **Launch Services:**
    ```sh
    docker-compose up --build
    ```
    This will build the necessary Docker images and start the LiveTran server, along with any dependent services.

## 🛠️ Tech Stack

- **Backend:** Go
- **Protocols:** SRT, WebRTC, HLS
- **Transcoding:** FFmpeg
- **Storage:** Cloudflare R2
- **Containerization:** Docker, Kubernetes
- **CI/CD:** GitHub Actions
- **Monitoring:** Prometheus, Grafana

## 🗺️ Project Roadmap

This project is developed in phases. Here’s a look at what’s done and what’s next:

### **Phase 1: MVP (Complete)** ✅

- [x] Core project setup (Go modules, Git)
- [x] SRT ingest endpoint
- [x] JWT-secured stream key generation
- [x] REST/gRPC API for stream management
- [ ] Basic user authentication and metadata storage
- [ ] Dockerfiles for all services

### **Phase 2: Optimization & Scalability (In Progress)** 🚧

- [ ] LL-HLS output for ultra-low latency
- [ ] WebRTC ingest support
- [ ] Multi-bitrate HLS transcoding (1080p, 720p, 480p)
- [ ] Kubernetes manifests and auto-scaling
- [ ] Cloudflare CDN integration with signed playback URLs
- [ ] Real-time analytics with Prometheus/Grafana

### **Phase 3: Production Polish (Planned)** 📝

- [ ] Infrastructure as Code (Terraform)
- [ ] Hardened security (rate limiting, encrypted HLS)
- [ ] Admin dashboard for stream and user management
- [ ] Multi-region deployment readiness
- [ ] Finalized documentation and CI/CD pipelines

For a more detailed breakdown, see the [TODO.md](temp/TODO.md) file.

## 📜 API Reference

The API allows you to manage streams programmatically. All request/response bodies are in JSON.

| Endpoint          | Method | Description                               |
| ----------------- | ------ | ----------------------------------------- |
| `/stream/start`   | `POST` | Starts a new stream and transcoding task. |
| `/stream/stop`    | `POST` | Stops an active stream.                   |
| `/stream/status`  | `POST` | Checks the current status of a stream.    |

**Example: Start a Stream**
```http
POST /stream/start
Content-Type: application/json

{
  "stream_id": "my-awesome-stream",
  "webhook_urls": ["https://my-service.com/webhook"]
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue to discuss your ideas.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.