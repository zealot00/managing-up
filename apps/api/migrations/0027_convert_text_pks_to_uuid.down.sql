-- Reverse of 0027: Convert skills.id and mcp_servers.id back from UUID to TEXT
-- This uses the id_migration_map table to restore original TEXT IDs.

-- -------------------------------------------------------
-- PHASE 1: skills.id UUID → TEXT
-- -------------------------------------------------------

-- Add back TEXT columns
ALTER TABLE skills ADD COLUMN IF NOT EXISTS id_text TEXT;

-- Restore from mapping
UPDATE skills SET id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND skills.id = m.new_uuid_id AND skills.id_text IS NULL;
ALTER TABLE skills ALTER COLUMN id_text SET NOT NULL;

-- Add TEXT columns to child tables
ALTER TABLE skill_versions ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE executions ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE replay_snapshots ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE skill_dependencies ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE skill_dependencies ADD COLUMN IF NOT EXISTS dependency_skill_id_text TEXT;
ALTER TABLE skill_ratings ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE skill_installs ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE skill_publish_approvals ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE skill_capability_snapshots ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE seh_releases ADD COLUMN IF NOT EXISTS skill_id_text TEXT;
ALTER TABLE mcp_server_permissions ADD COLUMN IF NOT EXISTS skill_id_text TEXT;

-- Populate from mapping
UPDATE skill_versions sv SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sv.skill_id = m.new_uuid_id AND sv.skill_id_text IS NULL;
UPDATE executions e SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND e.skill_id = m.new_uuid_id AND e.skill_id_text IS NULL;
UPDATE tasks t SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND t.skill_id = m.new_uuid_id AND t.skill_id_text IS NULL;
UPDATE replay_snapshots rs SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND rs.skill_id = m.new_uuid_id AND rs.skill_id_text IS NULL;
UPDATE skill_dependencies sd SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sd.skill_id = m.new_uuid_id AND sd.skill_id_text IS NULL;
UPDATE skill_dependencies sd SET dependency_skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sd.dependency_skill_id = m.new_uuid_id AND sd.dependency_skill_id_text IS NULL;
UPDATE skill_ratings sr SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sr.skill_id = m.new_uuid_id AND sr.skill_id_text IS NULL;
UPDATE skill_installs si SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND si.skill_id = m.new_uuid_id AND si.skill_id_text IS NULL;
UPDATE skill_publish_approvals spa SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND spa.skill_id = m.new_uuid_id AND spa.skill_id_text IS NULL;
UPDATE skill_capability_snapshots scs SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND scs.skill_id = m.new_uuid_id AND scs.skill_id_text IS NULL;
UPDATE seh_releases sr SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND sr.skill_id = m.new_uuid_id AND sr.skill_id_text IS NULL;
UPDATE mcp_server_permissions msp SET skill_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'skills' AND msp.skill_id = m.new_uuid_id AND msp.skill_id_text IS NULL;

-- Drop FK constraints
DO $$ BEGIN ALTER TABLE skill_versions DROP CONSTRAINT IF EXISTS skill_versions_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE executions DROP CONSTRAINT IF EXISTS executions_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS fk_skill_deps_dep; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS fk_skill_ratings_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_installs DROP CONSTRAINT IF EXISTS fk_skill_installs_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_publish_approvals DROP CONSTRAINT IF EXISTS fk_skill_pub_skill; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_capability_snapshots DROP CONSTRAINT IF EXISTS skill_capability_snapshots_skill_id_fkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies DROP CONSTRAINT IF EXISTS skill_dependencies_skill_id_dependency_skill_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings DROP CONSTRAINT IF EXISTS skill_ratings_skill_id_user_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Swap columns
ALTER TABLE skills DROP COLUMN IF EXISTS id;
ALTER TABLE skills RENAME COLUMN id_text TO id;
DO $$ BEGIN ALTER TABLE skills DROP CONSTRAINT IF EXISTS skills_pkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
ALTER TABLE skills ADD PRIMARY KEY (id);

ALTER TABLE skill_versions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_versions RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE executions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE executions RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS skill_id;
ALTER TABLE tasks RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE replay_snapshots DROP COLUMN IF EXISTS skill_id;
ALTER TABLE replay_snapshots RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE skill_dependencies DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_dependencies RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE skill_dependencies DROP COLUMN IF EXISTS dependency_skill_id;
ALTER TABLE skill_dependencies RENAME COLUMN dependency_skill_id_text TO dependency_skill_id;
ALTER TABLE skill_ratings DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_ratings RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE skill_installs DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_installs RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE skill_publish_approvals DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_publish_approvals RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE skill_capability_snapshots DROP COLUMN IF EXISTS skill_id;
ALTER TABLE skill_capability_snapshots RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE seh_releases DROP COLUMN IF EXISTS skill_id;
ALTER TABLE seh_releases RENAME COLUMN skill_id_text TO skill_id;
ALTER TABLE mcp_server_permissions DROP COLUMN IF EXISTS skill_id;
ALTER TABLE mcp_server_permissions RENAME COLUMN skill_id_text TO skill_id;

-- Re-add FK constraints with TEXT types
DO $$ BEGIN ALTER TABLE skill_versions ADD CONSTRAINT skill_versions_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE executions ADD CONSTRAINT executions_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE RESTRICT; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE tasks ADD CONSTRAINT tasks_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE SET NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT fk_skill_deps_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT fk_skill_deps_dep FOREIGN KEY (dependency_skill_id) REFERENCES skills(id) ON DELETE RESTRICT; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings ADD CONSTRAINT fk_skill_ratings_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_installs ADD CONSTRAINT fk_skill_installs_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_publish_approvals ADD CONSTRAINT fk_skill_pub_skill FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_capability_snapshots ADD CONSTRAINT skill_capability_snapshots_skill_id_fkey FOREIGN KEY (skill_id) REFERENCES skills(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_dependencies ADD CONSTRAINT skill_dependencies_skill_id_dependency_skill_id_key UNIQUE (skill_id, dependency_skill_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE skill_ratings ADD CONSTRAINT skill_ratings_skill_id_user_id_key UNIQUE (skill_id, user_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- -------------------------------------------------------
-- PHASE 2: mcp_servers.id UUID → TEXT
-- -------------------------------------------------------
ALTER TABLE mcp_servers ADD COLUMN IF NOT EXISTS id_text TEXT;
UPDATE mcp_servers SET id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mcp_servers.id = m.new_uuid_id AND mcp_servers.id_text IS NULL;
ALTER TABLE mcp_servers ALTER COLUMN id_text SET NOT NULL;

ALTER TABLE mcp_server_permissions ADD COLUMN IF NOT EXISTS mcp_server_id_text TEXT;
ALTER TABLE mcp_router_catalog ADD COLUMN IF NOT EXISTS server_id_text TEXT;
ALTER TABLE mcp_router_logs ADD COLUMN IF NOT EXISTS matched_server_id_text TEXT;
ALTER TABLE mcp_router_sync_log ADD COLUMN IF NOT EXISTS server_id_text TEXT;

UPDATE mcp_server_permissions msp SET mcp_server_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND msp.mcp_server_id = m.new_uuid_id AND msp.mcp_server_id_text IS NULL;
UPDATE mcp_router_catalog mrc SET server_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrc.server_id = m.new_uuid_id AND mrc.server_id_text IS NULL;
UPDATE mcp_router_logs mrl SET matched_server_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrl.matched_server_id = m.new_uuid_id AND mrl.matched_server_id_text IS NULL;
UPDATE mcp_router_sync_log mrsl SET server_id_text = m.old_text_id FROM id_migration_map m WHERE m.table_name = 'mcp_servers' AND mrsl.server_id = m.new_uuid_id AND mrsl.server_id_text IS NULL;

DO $$ BEGIN ALTER TABLE mcp_server_permissions DROP CONSTRAINT IF EXISTS fk_mcp_perms_server; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE mcp_router_catalog DROP CONSTRAINT IF EXISTS mcp_router_catalog_server_id_key; EXCEPTION WHEN OTHERS THEN NULL; END $$;

ALTER TABLE mcp_servers DROP COLUMN IF EXISTS id;
ALTER TABLE mcp_servers RENAME COLUMN id_text TO id;
DO $$ BEGIN ALTER TABLE mcp_servers DROP CONSTRAINT IF EXISTS mcp_servers_pkey; EXCEPTION WHEN OTHERS THEN NULL; END $$;
ALTER TABLE mcp_servers ADD PRIMARY KEY (id);

ALTER TABLE mcp_server_permissions DROP COLUMN IF EXISTS mcp_server_id;
ALTER TABLE mcp_server_permissions RENAME COLUMN mcp_server_id_text TO mcp_server_id;
ALTER TABLE mcp_router_catalog DROP COLUMN IF EXISTS server_id;
ALTER TABLE mcp_router_catalog RENAME COLUMN server_id_text TO server_id;
ALTER TABLE mcp_router_logs DROP COLUMN IF EXISTS matched_server_id;
ALTER TABLE mcp_router_logs RENAME COLUMN matched_server_id_text TO matched_server_id;
ALTER TABLE mcp_router_sync_log DROP COLUMN IF EXISTS server_id;
ALTER TABLE mcp_router_sync_log RENAME COLUMN server_id_text TO server_id;

DO $$ BEGIN ALTER TABLE mcp_server_permissions ADD CONSTRAINT fk_mcp_perms_server FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE mcp_router_catalog ADD CONSTRAINT mcp_router_catalog_server_id_key UNIQUE (server_id); EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- Clean up
DROP TABLE IF EXISTS id_migration_map;
