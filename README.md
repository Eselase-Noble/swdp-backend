backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/            # JWT, login, middleware
│   ├── users/           # user service
│   ├── projects/        # project service
│   ├── workspaces/      # workspace lifecycle
│   ├── files/           # file read/write abstraction
│   ├── execution/       # docker orchestration
│   ├── websocket/       # terminal & logs
│   ├── audit/           # audit logging
│   ├── middleware/      # auth, rbac, rate-limit
│   ├── config/          # env + config loading
│   └── database/        # db bootstrap
├── pkg/
│   ├── docker/          # docker client wrapper
│   ├── logger/          # structured logging
│   └── rbac/            # role enforcement
├── migrations/
├── Dockerfile
├── go.mod
└── go.sum
