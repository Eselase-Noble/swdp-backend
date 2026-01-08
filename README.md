# 🧩 SWDP Backend
**Web-Based Software Development Platform – Backend Service**

---

## 📌 Overview

The **SWDP Backend** is the core server-side component of the **Web-Based Software Development Platform (SWDP)**.  
It provides secure APIs, workspace orchestration, and real-time communication to enable developers to write, run, and manage code entirely in a browser-based environment without storing proprietary code on local machines.

This backend is designed with **scalability, security, and multi-tenancy** in mind and serves as the foundation for a cloud-hosted development platform.

---

## 🏗️ Core Responsibilities

- 🔐 Authentication & authorization
- 👥 User, project, and workspace management
- 🐳 Secure workspace execution using Docker
- 🧪 Project lifecycle management
- 🔄 Real-time communication via WebSockets (logs & terminal)
- 📜 Audit logging for traceability
- 🗄️ PostgreSQL-backed persistence

---

## 🛠️ Tech Stack

### Backend
- **Language:** Go (Golang)
- **Framework:** `chi` HTTP router
- **Database:** PostgreSQL
- **Driver:** `pgx`
- **Auth:** JWT-based authentication
- **WebSockets:** `github.com/coder/websocket`
- **Container Runtime:** Docker Engine API

### Infrastructure
- Docker (workspace execution)
- PostgreSQL extensions:
    - `pgcrypto`

---

## 📂 Project Structure

```text
.
├── cmd/
│   └── api/                  # Application entry point
│
├── internal/
│   ├── auth/                 # Authentication & JWT logic
│   ├── config/               # Application configuration
│   ├── execution/            # Docker container lifecycle management
│   ├── middleware/           # Auth, logging, request context
│   ├── projects/             # Project APIs
│   ├── users/                # User APIs
│   ├── workspaces/           # Workspace APIs
│   ├── websocket/            # Terminal & logs WebSocket handlers
│
├── migrations/
│   ├── 001_init.sql          # Database schema
│   └── 002_seed_data.sql     # Development seed data
│
├── go.mod
├── go.sum
└── README.md
```
# 🧠 Domain Model
## Key Entities

- **Users** – platform accounts (admin / developer)

- **Projects** – logical grouping of work

- **Project Members** – user roles per project

- **Workspaces** – isolated execution environments

- **Audit Logs** – system activity tracking


# 🗄️ Database Setup
## Requirements
- **PostgreSQL ≥ 13**
- **Required extensions:**
```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```
- 

## Run Migrations
```bash 
 psql -h localhost -U <db_user> -d <db_name> -f migrations/001_init.sql
```
````bash
 psql -h localhost -U <db_user> -d <db_name> -f migrations/002_seed_data.sql
````

## 🚀 Running the Backend
### Prerequisites

- Go ≥ 1.25

- Docker (running)

- PostgreSQL (running)

- Run Locally
```bash
  go mod tidy
```
```bash
  go run ./cmd/api
```



### The API will be available at:

- http://localhost:8282

## 🔌 API Overview
### Base Path
     - /api

#### Key Endpoints
####  **Authentication**

    POST /api/auth/login

    POST /api/auth/register

### **Projects**

    GET /api/projects/all

    POST /api/projects/create-project

    GET /api/projects/{id}

### **Workspaces**

    POST /api/projects/{projectID}/workspaces
    
    POST /api/workspaces/{id}/start
    
    POST /api/workspaces/{id}/stop

### **WebSockets**

    GET /api/ws/logs/{workspaceID}
    
    GET /api/ws/terminal/{workspaceID}

## 🧪 Health Check
    GET /health


### Response:

**{
"status": "ok"
}**

## 🔐 Security Model

- JWT-based authentication

- User context injected via middleware

- Role-based access control (RBAC-ready)

- Containers isolated per workspace

- No source code stored on developer machines

## 🧑‍💻 Development Notes

- Seed scripts are idempotent and safe to re-run

- Workspace containers are named deterministically:

 ```text 
 ws-<workspace_id>
 ```

- Designed for remote execution environments

- WebSocket connections support:

  - Interactive terminals
  
  - Live execution logs

## 🛣️ Roadmap

- ✅ Core API & workspace lifecycle

- ✅ Database migrations & seed data

- ⏳ Role-based access control

- ⏳ Rate limiting

- ⏳ Workspace resource quotas

- ⏳ CI/CD integration

- ⏳ Kubernetes support

- 📜 License

```text 
This project is proprietary and intended for internal or controlled deployment.
```
```text
All rights reserved.

👤 Author

Noble Eselase Vulley (Nobleson)

Software Engineer | Platform Architect
```