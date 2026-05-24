-- ============================================================
-- SWDP Development Seed — Known Test Credentials
-- Safe to re-run (ON CONFLICT DO NOTHING)
-- ============================================================

-- ──────────────────────────────────────────────────────────
-- ADMIN USERS
-- ──────────────────────────────────────────────────────────
INSERT INTO users (username, name, email, password_hash, role)
VALUES
    ('superadmin',
     'Super Admin',
     'admin@swdp.io',
     crypt('Admin@1234', gen_salt('bf')),
     'admin'),

    ('nobleson',
     'Noble Eselase',
     'nobleson@swdp.io',
     crypt('Noble@1234', gen_salt('bf')),
     'admin')
ON CONFLICT (email) DO NOTHING;

-- ──────────────────────────────────────────────────────────
-- DEVELOPER USERS
-- ──────────────────────────────────────────────────────────
INSERT INTO users (username, name, email, password_hash, role)
VALUES
    ('alice',
     'Alice Johnson',
     'alice@swdp.io',
     crypt('Alice@1234', gen_salt('bf')),
     'developer'),

    ('bob',
     'Bob Smith',
     'bob@swdp.io',
     crypt('Bob@1234', gen_salt('bf')),
     'developer'),

    ('carol',
     'Carol Williams',
     'carol@swdp.io',
     crypt('Carol@1234', gen_salt('bf')),
     'developer'),

    ('david',
     'David Brown',
     'david@swdp.io',
     crypt('David@1234', gen_salt('bf')),
     'developer')
ON CONFLICT (email) DO NOTHING;

-- ──────────────────────────────────────────────────────────
-- DEMO PROJECT (owned by superadmin)
-- ──────────────────────────────────────────────────────────
INSERT INTO projects (project_name, owner_id, version, environment)
SELECT
    'SWDP Demo',
    user_id,
    1,
    'dev'::environment_enum
FROM users
WHERE email = 'admin@swdp.io'
ON CONFLICT DO NOTHING;

-- ──────────────────────────────────────────────────────────
-- ADD ALICE AND BOB TO THE DEMO PROJECT
-- ──────────────────────────────────────────────────────────
INSERT INTO project_members (user_id, project_id, role)
SELECT
    u.user_id,
    p.project_id,
    'contributor'
FROM users u
CROSS JOIN projects p
WHERE u.email IN ('alice@swdp.io', 'bob@swdp.io')
  AND p.project_name = 'SWDP Demo'
ON CONFLICT DO NOTHING;
