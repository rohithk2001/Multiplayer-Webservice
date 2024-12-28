# Multiplayer Web Service

A backend service designed for managing multiplayer game modes, including tracking active users, managing game states, and providing analytics. Built using Go, MongoDB, Redis, and gRPC.

---

## Features

- **Game Mode Management:** Create, update, and manage game modes.
- **Active User Tracking:** Track active users in real-time across game modes.
- **Cache Layer:** Redis caching for faster responses.
- **MongoDB Storage:** Persistent storage for game mode data.
- **gRPC and REST API:** Supports gRPC calls and REST endpoints.

---

## Technologies Used

- **Go:** Primary programming language.
- **MongoDB:** Database for storing game mode data.
- **Redis:** Cache layer for improving performance.
- **gRPC:** High-performance RPC framework.
- **Docker:** Containerization for deployment.

---

## Setup Instructions

### Prerequisites

- [Docker](https://www.docker.com/)
- [Go](https://golang.org/) (for local development)

### Clone the Repository

```bash
git clone https://github.com/rohithk2001/Multiplayer-Webservice.git
cd <Multiplayer-Webservice>
```

#Environment Variables

Create a `.env` file and add the following:

```env
MONGODB_URI=mongodb://mongo:27017/multiplayer_db
REDIS_ADDR=redis:6379
REDIS_PASS=<your-redis-password>
SERVER_PORT=8080
GRPC_PORT=50051
```

## Run the Application

### Using Docker Compose

Build and start the services:

```bash



docker-compose up --build
```
## Contributing

Contributions are welcome! Please fork the repository and submit a pull request for any improvements or features you'd like to add.


## i will upload a detailed Readmefile very soon sorry for inconvenience
