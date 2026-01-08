CREATE EXTENSION IF NOT EXISTS pgcrypto;


-- TRUNCATE TABLE
--     audit_logs,
--     workspaces,
--     project_members,
--     projects,
--     users
-- RESTART IDENTITY CASCADE;


-------------------------------------------------
-- USERS
-------------------------------------------------
CREATE TABLE users (
                       user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       username TEXT UNIQUE NOT NULL ,
                       name TEXT NOT NULL ,
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
-- 1️⃣ Create the enum type for environments
CREATE TYPE environment_enum AS ENUM ('dev', 'staging', 'prod');

-- 2️⃣ Create the projects table
CREATE TABLE projects (
                          project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          project_name TEXT NOT NULL,
                          description TEXT,
                          owner_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
                          version INT,
                          environment environment_enum NOT NULL DEFAULT 'dev',  -- enum with default

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
                                 user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
                                 project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
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
                            project_id UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
                            user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
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
