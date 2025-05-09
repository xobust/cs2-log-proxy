# CS2 log proxy (WIP)

A Go + React application to receive, proxy, and stream CS2 log packages in real time, with a management UI.

## Features

- Receives CS2 log packages and forwards to multiple receivers
- Real-time log streaming via WebSocket
- File-based log storage (S3 planned)
- Management UI (React + Material-UI)
- Configuration management

## Getting Started

### Backend (Go)

```sh
cd cs2-log-manager
go run main.go
```

### Frontend (React)

```sh
cd cs2-log-manager/web
npm install
npm start
```

- Backend runs on `localhost:8081`
- Frontend runs on `localhost:3000` (proxy to backend)

## Status

Work in progress. Contributions and feedback welcome!
