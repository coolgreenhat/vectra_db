# VectraDB

A production-grade vector database built in Go, designed for high-performance similarity search and hybrid search capabilities.

## Features

- **Vector Storage**: Store and manage high-dimensional vectors with metadata
- **Similarity Search**: Perform cosine similarity search with filtering
- **Hybrid Search**: Combine vector similarity with keyword search using BM25
- **Document Management**: Store and manage documents with tags
- **RESTful API**: Clean, well-documented REST API
- **Production Ready**: Structured logging, error handling, middleware, and graceful shutdown
- **Persistent Storage**: Built on BoltDB for reliable data persistence
- **Configurable**: Environment-based configuration
- **Health Checks**: Built-in health monitoring endpoints

## Architecture

```
├── cmd/                    # Application entry points
│   └── vectordbd/         # Main server application
├── internal/              # Private application code
│   ├── api/               # HTTP handlers and routing
│   ├── config/            # Configuration management
│   ├── logger/            # Structured logging
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models and DTOs
│   ├── store/             # Data storage interfaces and implementations
│   └── utils/             # Utility functions
├── pkg/                   # Public library code
│   ├── errors/            # Error handling
│   └── response/          # HTTP response utilities
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
├── deployments/           # Deployment configurations
└── tests/                 # Test files
```

## Quick Start

### Prerequisites

- Go 1.24.7 or later
- Make (optional, for using Makefile)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd vectra_db
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
make build
# or
go build -o bin/vectordbd ./cmd/vectordbd
```

4. Run the server:
```bash
make run
# or
./bin/vectordbd
```

The server will start on port 8080 by default.

## Configuration

VectraDB can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `vectra.db` | Database file path |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |
| `READ_TIMEOUT` | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `120s` | HTTP idle timeout |
| `DB_TIMEOUT` | `1s` | Database operation timeout |

## API Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Vector Operations

#### Create Vector
```http
POST /vectors
Content-Type: application/json

{
  "id": "vector-1",
  "vector": [0.1, 0.2, 0.3, 0.4],
  "text": "Sample text",
  "metadata": {
    "category": "example",
    "source": "test"
  }
}
```

#### Get Vector
```http
GET /vectors/{id}
```

#### Update Vector
```http
PUT /vectors/{id}
Content-Type: application/json

{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "text": "Updated text",
  "metadata": {
    "category": "updated"
  }
}
```

#### Delete Vector
```http
DELETE /vectors/{id}
```

#### List Vectors
```http
GET /vectors?limit=10&offset=0
```

### Search Operations

#### Vector Search
```http
POST /search
Content-Type: application/json

{
  "query": [0.1, 0.2, 0.3, 0.4],
  "top_k": 10,
  "limit": 10,
  "page": 1,
  "filter": {
    "category": "example"
  },
  "weights": {
    "vector": 1.0,
    "metadata": 0.0
  }
}
```

#### Hybrid Search
```http
POST /search/hybrid
Content-Type: application/json

{
  "query": "search text",
  "query_vector": [0.1, 0.2, 0.3, 0.4],
  "vector_weight": 0.5,
  "keyword_weight": 0.5,
  "limit": 10,
  "page": 1
}
```

### Document Operations

#### Create Document
```http
POST /documents
Content-Type: application/json

{
  "id": "doc-1",
  "title": "Sample Document",
  "content": "Document content here",
  "tags": ["tag1", "tag2"]
}
```

#### Get Document
```http
GET /documents/{id}
```

#### Update Document
```http
PUT /documents/{id}
Content-Type: application/json

{
  "title": "Updated Document",
  "content": "Updated content",
  "tags": ["tag1", "tag3"]
}
```

#### Delete Document
```http
DELETE /documents/{id}
```

#### List Documents
```http
GET /documents?limit=10&offset=0
```

#### List Documents by Tag
```http
GET /documents/tags/{tag}?limit=10&offset=0
```

### Health Check

#### Health Status
```http
GET /health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "healthy"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Development

### Running Tests
```bash
make test
# or
go test ./... -v
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

## Docker

### Build Docker Image
```bash
docker build -t vectradb .
```

### Run with Docker
```bash
docker run -p 8080:8080 vectradb
```

### Docker Compose
```bash
docker-compose up
```

## Production Deployment

### Environment Variables
Set the following environment variables for production:

```bash
export PORT=8080
export DB_PATH=/data/vectra.db
export LOG_LEVEL=info
export LOG_FORMAT=json
export READ_TIMEOUT=30s
export WRITE_TIMEOUT=30s
export IDLE_TIMEOUT=120s
```

### Health Monitoring
The application provides a health check endpoint at `/health` that can be used by load balancers and monitoring systems.

### Graceful Shutdown
The application supports graceful shutdown on SIGINT and SIGTERM signals, allowing up to 30 seconds for ongoing requests to complete.

## Performance Considerations

- **Vector Dimensions**: Supports vectors up to 10,000 dimensions
- **Batch Operations**: Use pagination for large result sets
- **Memory Usage**: Vectors are cached in memory for fast access
- **Database**: BoltDB provides ACID transactions and crash recovery

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue on GitHub.
