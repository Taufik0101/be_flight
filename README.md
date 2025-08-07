## Quick Start

### Clone Repository

```bash
git clone https://github.com/Taufik0101/be_flight.git
cd be_flight
```

### Install Dependencies
```bash
go mod tidy
go get .
```

### Run Docker Compose
```bash
Make Sure Local Docker already run
docker-compose up --build
```

### How to use
UI
```bash
http://localhost:3000
```

### OR

Postman
1. http://localhost:3000/api/flights/search -> POST

Payload
```json
{
    "from": "CGK",
    "to": "DPS",
    "date": "2025-07-10",
    "passengers": 2
}
```

Response
```json
{
    "data": {
        "search_id": "uuid",
        "status": "processing"
    },
    "message": "Search request submitted",
    "success": true
}
```

2. http://localhost:3000/api/flights/search/{{search_id}}/stream -> GET

OR

```bash
go run main.go (inside root folder)