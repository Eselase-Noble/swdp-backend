CREATE EXTENSION IF NOT EXISTS pgcrypto;

-------------------------------------------------
-- USERS
-------------------------------------------------
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       role TEXT NOT NULL,

                       created_by UUID,
                       updated_by UUID,
                       created_at TIMESTAMP NOT NULL DEFAULT now(),
                       updated_at TIMESTAMP NOT NULL DEFAULT now(),
                       deleted_yn BOOLEAN NOT NULL DEFAULT false
);

-------------------------------------------------
-- PROJECTS
-------------------------------------------------
CREATE TABLE projects (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name TEXT NOT NULL,
                          owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                          created_by UUID,
                          updated_by UUID,
                          created_at TIMESTAMP NOT NULL DEFAULT now(),
                          updated_at TIMESTAMP NOT NULL DEFAULT now(),
                          deleted_yn BOOLEAN NOT NULL DEFAULT false
);

-------------------------------------------------
-- PROJECT MEMBERS
-------------------------------------------------
CREATE TABLE project_members (
                                 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                 project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                                 role TEXT NOT NULL,

                                 created_by UUID,
                                 updated_by UUID,
                                 created_at TIMESTAMP NOT NULL DEFAULT now(),
                                 updated_at TIMESTAMP NOT NULL DEFAULT now(),
                                 deleted_yn BOOLEAN NOT NULL DEFAULT false,

                                 PRIMARY KEY (user_id, project_id)
);

-------------------------------------------------
-- WORKSPACES
-------------------------------------------------
CREATE TABLE workspaces (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            status TEXT NOT NULL,

                            created_by UUID,
                            updated_by UUID,
                            created_at TIMESTAMP NOT NULL DEFAULT now(),
                            updated_at TIMESTAMP NOT NULL DEFAULT now(),
                            deleted_yn BOOLEAN NOT NULL DEFAULT false
);

-------------------------------------------------
-- AUDIT LOGS
-------------------------------------------------
CREATE TABLE audit_logs (
                            id BIGSERIAL PRIMARY KEY,
                            user_id UUID,
                            action TEXT NOT NULL,
                            resource TEXT NOT NULL,

                            created_by UUID,
                            updated_by UUID,
                            created_at TIMESTAMP NOT NULL DEFAULT now(),
                            updated_at TIMESTAMP NOT NULL DEFAULT now(),
                            deleted_yn BOOLEAN NOT NULL DEFAULT false
);
