# Remote Monitoring and Management (RMM)

A lightweight platform to connect, monitor, and dispatch jobs to remote agents.

## Getting Started

### 1. Configure the Database

Ensure PostgreSQL is running and set the connection string environment variable.

```bash
# Example
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
```

### 2. Run Database Migrations

Apply the repository migrations before starting the backend.

```bash
go run ./cmd/migrate
```

### 3. Start the Backend Server

Run the Go server from the root directory. This will start the API and WebSocket listeners.

```bash
go run ./cmd/server
```
*(Available by default at `https://localhost:8081`)*

### 4. Start the Frontend UI

Open a second terminal, navigate into the `ui` folder, and start the web dashboard.

```bash
cd ui
npm install
npm run dev
```
*(Visit the provided Vite localhost URL in your browser)*

### 5. Connect an Agent

Open a third terminal and run the agent binary. The agent will automatically authenticate and connect to the backend server, instantly appearing as "Online" in your UI dashboard.

```bash
go run ./cmd/agent
```
