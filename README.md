# CS2 log proxy (WIP)

A Go + React application to receive CS2 log packages (logaddress_add_http) and store them in a file system.
It also provides a management UI to view and manage the logs with real-time updates.
Proxying logs to multiple receivers is planned.

## Roadmap

- [x] Implement log receiving
- [x] Implement robust chunk reconstruction
- [ ] Implement robust log proxying
- [ ] Add support for log delay
- [ ] Implement log storage (S3 planned)
- [ ] Implement management UI (React + Material-UI)
- [ ] Run distributed on multiple machines

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
