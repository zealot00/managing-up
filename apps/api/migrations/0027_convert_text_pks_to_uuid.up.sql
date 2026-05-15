-- ============================================================
-- Migration 0027: Convert skills.id and mcp_servers.id from TEXT to UUID
--
-- Strategy:
--   1. Create id_migration_map to preserve old TEXT → new UUID mapping
--   2. Add new UUID column (_uuid) to each parent table
--   3. Generate deterministic UUIDs from existing TEXT ids (md5-based v4)
--   4. Save mapping into id_migration_map
--   5. Add _uuid columns to all child tables, populate from mapping
--   6. Drop old FK constraints, drop old TEXT columns, rename _uuid columns
--   7. Re-add FK constraints with UUID types
--   8. Re-create dependent objects (indexes, unique constraints)
-- ============================================================

-- -------------------------------------------------------
-- PHASE 0: Migration map table (for Go code to look up old→new IDs)
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS id_migration_map (
    table_name  TEXT NOT NULL,
    old_text_id TEXT NOT NULL,
    new_uuid_id UUID NOT NULL,
    PRIMARY KEY (table_name, old_text_id)
);

-- -------------------------------------------------------
-- PHASE 1: skills.id TEXT → UUID
-- -------------------------------------------------------

-- 1a. Add UUID column to skills
ALTER TABLE skills ADD COLUMN IF NOT EXISTS id_uuid UUID;

-- 1b. Generate deterministic UUIDs and store mapping
-- Use md5 of a namespaced string, formatted as UUID v4 variant
WITH map AS (
    INSERT INTO id_migration_map (table_name, old_text_id, new_uuid_id)
    SELECT
        'skills',
        id,
        (substr(m, 1, 8) || '-' || substr(m, 9, 4) || '-4' || substr(m, 13, 3) || '-' ||
         substr(m, 17, 4) || '-' || substr(m, 21, 12))::uuid
    FROM skills, LATERAL (SELECT md5('skill:' || skills.id) AS m) t
    WHERE id_uuid IS NULL
    ON CONFLICT (table_name, old_text_id) DO UPDATE SET new_uuid_id = EXCLUDED.new_uuid_id
    RETURNING old_text_id, new_uuid_id
)
UPDATE skills SET id_uuid = map.new_uuid_id
FROM map WHERE skills.id = map.old_text_id AND skills.id_uuid IS NULL;

-- Make id_uuid NOT NULL
UPDATE skills SET id_uuid = (SELECT new_uuid_id FROM id_migration_map WHERE table_name = 'skills' AND old_text_id = skills.id)
WHERE id_uuid IS NULL;
ALTER TABLE skills ALTER COLUMN id_uuid SET NOT NULL;

-- 1c. Add _uuid columns to all child tables
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE executions ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE replay_snapshots ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE skill_dependencies ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE skill_dependencies ADD COLUMN IF NOT EXISTS dependency_skill_id_uuid UUID;
ALTER TABLE skill_ratings ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE skill_installs ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE skill_publish_approvals ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE skill_capability_snapshots ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE seh_releases ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;
ALTER TABLE mcp_server_permissions ADD COLUMN IF NOT EXISTS skill_id_uuid UUID;

-- 1d. Populate child _uuid from mapping table
UPDATE skill_versions sv SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sv.skill_id = m.old_text_id AND sv.skill_id_uuid IS NULL;
UPDATE executions e SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND e.skill_id = m.old_text_id AND e.skill_id_uuid IS NULL;
UPDATE tasks t SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND t.skill_id = m.old_text_id AND t.skill_id_uuid IS NULL;
UPDATE replay_snapshots rs SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND rs.skill_id = m.old_text_id AND rs.skill_id_uuid IS NULL;
UPDATE skill_dependencies sd SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sd.skill_id = m.old_text_id AND sd.skill_id_uuid IS NULL;
UPDATE skill_dependencies sd SET dependency_skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sd.dependency_skill_id = m.old_text_id AND sd.dependency_skill_id_uuid IS NULL;
UPDATE skill_ratings sr SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sr.skill_id = m.old_text_id AND sr.skill_id_uuid IS NULL;
UPDATE skill_installs si SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND si.skill_id = m.old_text_id AND si.skill_id_uuid IS NULL;
UPDATE skill_publish_approvals spa SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND spa.skill_id = m.old_text_id AND spa.skill_id_uuid IS NULL;
UPDATE skill_capability_snapshots scs SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND scs.skill_id = m.old_text_id AND scs.skill_id_uuid IS NULL;
UPDATE seh_releases sr SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sr.skill_id = m.old_text_id AND sr.skill_id_uuid IS NULL;
UPDATE mcp_server_permissions msp SET skill_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'skills' AND msp.skill_id = m.old_text_id AND msp.skill_id_uuid IS NULL;

-- 1e. Drop all FK constraints referencing skills.id
DO $$ BEGIN ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS skill_versions_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE executions DROP CONSTRAINT IF EXISTS executions_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_dep; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS skill_dependencies_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS skill_dependencies_dependency_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS skill_ratings_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS skill_installs_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS skill_publish_approvals_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_capability_snapshots DROP CONSTRAINT IF EXISTS skill_capability_snapshots_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Drop unique constraints involving skill_id
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS skill_dependencies_skill_id_dependency_skill_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS skill_ratings_skill_id_user_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Drop indexes on old skill_id columns (they'll be recreated after rename)
DROP INDEX IF EXISTS idx_skill_versions_skill_id;
DROP INDEX IF EXISTS idx_skill_versions_skill;
DROP INDEX IF EXISTS idx_tasks_skill_id;
DROP INDEX IF EXISTS idx_skill_deps_skill;
DROP INDEX IF EXISTS idx_skill_deps_dep;
DROP INDEX IF EXISTS idx_skill_ratings_skill;
DROP INDEX IF EXISTS idx_skill_installs_skill;
DROP INDEX IF EXISTS idx_skill_pub_skill;
DROP INDEX IF EXISTS idx_snapshots_skill;
DROP INDEX IF EXISTS idx_mcp_permissions_skill;
DROP INDEX IF EXISTS idx_skills_status;

-- 1f. Swap columns: drop old TEXT, rename UUID
ALTER TABLE skills DROP COLUMN IF EXISTS id;
ALTER TABLE skills RENAME COLUMN id_uuid TO id;
DO $$ BEGIN ALTER TABLE skills DROP CONSTRAINT IF EXISTS skills_pkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
ALTER TABLE skills ADD PRIMARY KEY (id);
-- Restore name unique constraint (name column still exists)
DO $$ BEGIN ALTER TABLE skills ADD CONSTRAINT skills_name_key UNIQUE (name); EXCEPTION WHEN OTHERS THEN NULL; END $$;

ALTER TABLE skill_versions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_versions RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE executions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE executions RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE tasks DROP COLUMN IF EXISTS skill_id;
ALTER TABLE tasks RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE replay_snapshots DROP COLUMN IF EXISTS skill_id;
ALTER TABLE replay_snapshots RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE skill_dependencies DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_dependencies RENAME COLUMN skill_id_uuid TO skill_id;
ALTER TABLE skill_dependencies DROP COLUMN IF EXISTS dependency_skill_id;
ALTER TABLE skill_dependencies RENAME COLUMN dependency_skill_id_uuid TO dependency_skill_id;

ALTER TABLE skill_ratings DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_ratings RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE skill_installs DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_installs RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE skill_publish_approvals DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_publish_approvals RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE skill_capability_snapshots DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_capability_snapshots RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE seh_releases DROP COLUMN IF EXISTS skill_id;
ALTER TABLE seh_releases RENAME COLUMN skill_id_uuid TO skill_id;

ALTER TABLE mcp_server_permissions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE mcp_server_permissions RENAME COLUMN skill_id_uuid TO skill_id;

-- 1g. Re-add FK constraints with UUID types
DO $$ BEGIN ALTER TABLE skill_versions ADD CONSTRAINT skill_versions_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE executions ADD CONSTRAINT executions_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE RESTRICT; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE tasks ADD CONSTRAINT tasks_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE SET NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT fk_skill_deps_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT fk_skill_deps_dep FOREIGN KEY (dependency_skill_id) REFERENCES skills(id) ON DELETE RESTRICT; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings ADD CONSTRAINT fk_skill_ratings_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_installs ADD CONSTRAINT fk_skill_installs_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_publish_approvals ADD CONSTRAINT fk_skill_pub_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_capability_snapshots ADD CONSTRAINT skill_capability_snapshots_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Re-add unique constraints
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT skill_dependencies_skill_id_dependency_skill_id_key UNIQUE (skill_id, dependency_skill_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings ADD CONSTRAINT skill_ratings_skill_id_user_id_key UNIQUE (skill_id, user_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Re-create indexes
CREATE INDEX IF NOT EXISTS idx_skill_versions_skill_id ON skill_versions(skill_id);
CREATE INDEX IF NOT EXISTS idx_tasks_skill_id ON tasks(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_deps_skill ON skill_dependencies(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_deps_dep ON skill_dependencies(dependency_skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_ratings_skill ON skill_ratings(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_installs_skill ON skill_installs(skill_id);
CREATE INDEX IF NOT EXISTS idx_skill_pub_skill ON skill_publish_approvals(skill_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_skill ON skill_capability_snapshots(skill_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_mcp_permissions_skill ON mcp_server_permissions(skill_id);
CREATE INDEX IF NOT EXISTS idx_skills_status ON skills(status);

-- -------------------------------------------------------
-- PHASE 2: mcp_servers.id TEXT → UUID
-- -------------------------------------------------------

-- 2a. Add UUID column to mcp_servers
ALTER TABLE mcp_servers ADD COLUMN IF NOT EXISTS id_uuid UUID;

-- 2b. Generate deterministic UUIDs and store mapping
WITH map AS (
    INSERT INTO id_migration_map (table_name, old_text_id, new_uuid_id)
    SELECT
        'mcp_servers',
        id,
        (substr(m, 1, 8) || '-' || substr(m, 9, 4) || '-4' || substr(m, 13, 3) || '-' ||
         substr(m, 17, 4) || '-' || substr(m, 21, 12))::uuid
    FROM mcp_servers, LATERAL (SELECT md5('mcp_server:' || mcp_servers.id) AS m) t
    WHERE id_uuid IS NULL
    ON CONFLICT (table_name, old_text_id) DO UPDATE SET new_uuid_id = EXCLUDED.new_uuid_id
    RETURNING old_text_id, new_uuid_id
)
UPDATE mcp_servers SET id_uuid = map.new_uuid_id
FROM map WHERE mcp_servers.id = map.old_text_id AND mcp_servers.id_uuid IS NULL;

UPDATE mcp_servers SET id_uuid = (SELECT new_uuid_id FROM id_migration_map WHERE table_name = 'mcp_servers' AND old_text_id = mcp_servers.id)
WHERE id_uuid IS NULL;
ALTER TABLE mcp_servers ALTER COLUMN id_uuid SET NOT NULL;

-- 2c. Add _uuid columns to child tables
ALTER TABLE mcp_server_permissions ADD COLUMN IF NOT EXISTS mcp_server_id_uuid UUID;
ALTER TABLE mcp_router_catalog ADD COLUMN IF NOT EXISTS server_id_uuid UUID;
ALTER TABLE mcp_router_logs ADD COLUMN IF NOT EXISTS matched_server_id_uuid UUID;
ALTER TABLE mcp_router_sync_log ADD COLUMN IF NOT EXISTS server_id_uuid UUID;

-- 2d. Populate child _uuid from mapping
UPDATE mcp_server_permissions msp SET mcp_server_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND msp.mcp_server_id = m.old_text_id AND msp.mcp_server_id_uuid IS NULL;
UPDATE mcp_router_catalog mrc SET server_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrc.server_id = m.old_text_id AND mrc.server_id_uuid IS NULL;
UPDATE mcp_router_logs mrl SET matched_server_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrl.matched_server_id = m.old_text_id AND mrl.matched_server_id_uuid IS NULL;
UPDATE mcp_router_sync_log mrsl SET server_id_uuid = m.new_uuid_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrsl.server_id = m.old_text_id AND mrsl.server_id_uuid IS NULL;

-- 2e. Drop FK constraints
DO $$ BEGIN ALTER TABLE mcp_server_permissions DROP CONSTRAINT IF EXISTS fk_mcp_perms_server; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE mcp_server_permissions DROP CONSTRAINT IF EXISTS mcp_server_permissions_mcp_server_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE mcp_router_catalog DROP CONSTRAINT IF EXISTS mcp_router_catalog_server_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Drop indexes on old columns
DROP INDEX IF EXISTS idx_mcp_permissions_server;
DROP INDEX IF EXISTS idx_mcp_catalog_server_id;
DROP INDEX IF EXISTS idx_mcp_router_logs_server;
DROP INDEX IF EXISTS idx_mcp_sync_log_server;

-- 2f. Swap columns
ALTER TABLE mcp_servers DROP COLUMN IF EXISTS id;
ALTER TABLE mcp_servers RENAME COLUMN id_uuid TO id;
DO $$ BEGIN ALTER TABLE mcp_servers DROP CONSTRAINT IF EXISTS mcp_servers_pkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
ALTER TABLE mcp_servers ADD PRIMARY KEY (id);
-- Restore name unique constraint
DO $$ BEGIN ALTER TABLE mcp_servers ADD CONSTRAINT mcp_servers_name_key UNIQUE (name); EXCEPTION WHEN OTHERS THEN NULL; END $$;

ALTER TABLE mcp_server_permissions DROP COLUMN IF EXISTS mcp_server_id;
ALTER TABLE mcp_server_permissions RENAME COLUMN mcp_server_id_uuid TO mcp_server_id;

ALTER TABLE mcp_router_catalog DROP COLUMN IF EXISTS server_id;
ALTER TABLE mcp_router_catalog RENAME COLUMN server_id_uuid TO server_id;

ALTER TABLE mcp_router_logs DROP COLUMN IF EXISTS matched_server_id;
ALTER TABLE mcp_router_logs RENAME COLUMN matched_server_id_uuid TO matched_server_id;

ALTER TABLE mcp_router_sync_log DROP COLUMN IF EXISTS server_id;
ALTER TABLE mcp_router_sync_log RENAME COLUMN server_id_uuid TO server_id;

-- 2g. Re-add FK constraints
DO $$ BEGIN ALTER TABLE mcp_server_permissions ADD CONSTRAINT fk_mcp_perms_server FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE mcp_router_catalog ADD CONSTRAINT mcp_router_catalog_server_id_key UNIQUE (server_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Re-create indexes
CREATE INDEX IF NOT EXISTS idx_mcp_permissions_server ON mcp_server_permissions(mcp_server_id);
CREATE INDEX IF NOT EXISTS idx_mcp_catalog_server_id ON mcp_router_catalog(server_id);
CREATE INDEX IF NOT EXISTS idx_mcp_router_logs_server ON mcp_router_logs(matched_server_id);
CREATE INDEX IF NOT EXISTS idx_mcp_sync_log_server ON mcp_router_sync_log(server_id);
