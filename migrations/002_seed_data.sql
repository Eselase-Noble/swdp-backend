-- ============================================
-- USERS (20)
-- ============================================
INSERT INTO users (
    user_id, username, name, email, password_hash, role
)
SELECT
    gen_random_uuid(),
    'user' || g,
    'User ' || g,
    'user' || g || '@example.com',
    crypt('password' || g, gen_salt('bf')),
    CASE WHEN g % 5 = 0 THEN 'admin' ELSE 'developer' END
FROM generate_series(1,20) g;

-- ============================================
-- PROJECTS (20)
-- ============================================
INSERT INTO projects (
    project_id, project_name, owner_id, version, environment
)
SELECT
    gen_random_uuid(),
    'Project ' || g,
    u.user_id,
    1,
    (
        CASE
            WHEN g % 3 = 0 THEN 'prod'
            WHEN g % 2 = 0 THEN 'staging'
            ELSE 'dev'
            END
        )::environment_enum
FROM generate_series(1,20) g
         JOIN LATERAL (
    SELECT user_id
    FROM users
    ORDER BY created_at
    OFFSET (g - 1) % 20
    LIMIT 1
    ) u ON true;

-- ============================================
-- PROJECT MEMBERS (20, GUARANTEED UNIQUE)
-- ============================================
WITH numbered_users AS (
    SELECT user_id, ROW_NUMBER() OVER (ORDER BY created_at) rn
    FROM users
),
     numbered_projects AS (
         SELECT project_id, ROW_NUMBER() OVER (ORDER BY created_at) rn
         FROM projects
     )
INSERT INTO project_members (
    user_id, project_id, role
)
SELECT
    u.user_id,
    p.project_id,
    CASE
        WHEN u.rn % 4 = 0 THEN 'maintainer'
        ELSE 'contributor'
        END
FROM numbered_users u
         JOIN numbered_projects p
              ON u.rn = p.rn
    ON CONFLICT DO NOTHING;

-- ============================================
-- WORKSPACES (20)
-- ============================================
INSERT INTO workspaces (
    id, project_id, user_id, status
)
SELECT
    gen_random_uuid(),
    p.project_id,
    u.user_id,
    CASE
        WHEN g % 3 = 0 THEN 'running'
        WHEN g % 2 = 0 THEN 'stopped'
        ELSE 'created'
        END
FROM generate_series(1,20) g
         JOIN LATERAL (
    SELECT user_id FROM users ORDER BY created_at OFFSET (g - 1) % 20 LIMIT 1
    ) u ON true
    JOIN LATERAL (
    SELECT project_id FROM projects ORDER BY created_at OFFSET (g - 1) % 20 LIMIT 1
    ) p ON true;

-- ============================================
-- AUDIT LOGS (20)
-- ============================================
INSERT INTO audit_logs (
    user_id, action, resource
)
SELECT
    u.user_id,
    CASE
        WHEN g % 3 = 0 THEN 'DELETE'
        WHEN g % 2 = 0 THEN 'UPDATE'
        ELSE 'CREATE'
        END,
    CASE
        WHEN g % 2 = 0 THEN 'PROJECT'
        ELSE 'WORKSPACE'
        END
FROM generate_series(1,20) g
         JOIN LATERAL (
    SELECT user_id FROM users ORDER BY created_at OFFSET (g - 1) % 20 LIMIT 1
    ) u ON true;
