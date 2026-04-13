# Remote Monitoring and Management (RMM)

A lightweight platform to connect, monitor, and dispatch jobs to remote agents.

## Getting Started

### 1. Configure the Database

Ensure PostgreSQL is running and set the connection string environment variable.

```bash
# Example
export DB_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
```

### 2. Start the Backend Server

Run the Go server from the root directory. This will start the API and WebSocket listeners.

```bash
go run cmd/server/main.go
```
*(Available by default at `http://localhost:8080`)*

### 3. Start the Frontend UI

Open a second terminal, navigate into the `ui` folder, and start the web dashboard.

```bash
cd ui
npm install
npm run dev
```
*(Visit the provided Vite localhost URL in your browser)*

### 4. Connect an Agent

Open a third terminal and run the agent binary. The agent will automatically authenticate and connect to the backend server, instantly appearing as "Online" in your UI dashboard.

```bash
go run cmd/agent/main.go
```
