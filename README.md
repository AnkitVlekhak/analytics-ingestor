# Analytics Ingestor

A high-throughput, low-latency gRPC service designed to ingest log metrics, buffer them for performance, and persist them into TimescaleDB.

## 📚 Documentation

- [**High Level Design (HLD)**](docs/HLD.md): System architecture, design goals, and context diagram.
- [**Low Level Design (LLD)**](docs/LLD.md): Code structure, database schema, and sequence diagrams.

## 🚀 Getting Started

### Prerequisites

- Go (Golang) 1.20+
- Docker & Docker Compose
- Make

### Running the Application

1.  **Start the Database**

    ```bash
    make db-up
    ```

2.  **Run the Server** (in a new terminal)

    ```bash
    make run-server
    ```

3.  **Run the Load Test Client** (in another terminal)
    ```bash
    make run-client
    ```

### Other Commands

- `make db-down`: Stop the database containers.
- `make db-logs`: View database logs.
- `make clean`: Tidy up modules.
- `make help`: Show all available commands.

## 📂 Project Structure

```
├── cmd/                # Entrypoints for applications
│   ├── server/         # gRPC Server main
│   └── client/         # Load Test Client
├── internal/           # Private application code
│   ├── app/            # Dependency Injection
│   ├── service/        # Business Logic (Ingestor)
│   └── repository/     # Data Access (TimescaleDB)
├── pkg/                # Public/Shared code
│   ├── api/v1/         # Generated gRPC Protobuf
│   └── tm/             # Transaction Manager
├── docs/               # Architecture Documentation
└── docker-compose.yml  # Infrastructure definition
```
