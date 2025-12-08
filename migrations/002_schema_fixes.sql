-- Migration: 002_schema_fixes
-- Description: Fix foreign key constraints, add missing columns, and add constraints
-- Created: 2025-12-07

-- Add payload columns to confluence_sync_log if they don't exist
ALTER TABLE confluence_sync_log ADD COLUMN IF NOT EXISTS request_payload JSONB;
ALTER TABLE confluence_sync_log ADD COLUMN IF NOT EXISTS response_payload JSONB;

-- Fix ON DELETE behavior for user foreign keys

-- prd_templates.created_by - SET NULL when user is deleted
ALTER TABLE prd_templates DROP CONSTRAINT IF EXISTS prd_templates_created_by_fkey;
ALTER TABLE prd_templates ADD CONSTRAINT prd_templates_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- prd_versions.changed_by - SET NULL when user is deleted
ALTER TABLE prd_versions DROP CONSTRAINT IF EXISTS prd_versions_changed_by_fkey;
ALTER TABLE prd_versions ADD CONSTRAINT prd_versions_changed_by_fkey
    FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL;

-- priority_scores.scored_by - SET NULL when user is deleted
ALTER TABLE priority_scores DROP CONSTRAINT IF EXISTS priority_scores_scored_by_fkey;
ALTER TABLE priority_scores ADD CONSTRAINT priority_scores_scored_by_fkey
    FOREIGN KEY (scored_by) REFERENCES users(id) ON DELETE SET NULL;

-- discovery_items.created_by - SET NULL when user is deleted
ALTER TABLE discovery_items DROP CONSTRAINT IF EXISTS discovery_items_created_by_fkey;
ALTER TABLE discovery_items ADD CONSTRAINT discovery_items_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

-- prds.owner_id - RESTRICT deletion to prevent orphan PRDs
ALTER TABLE prds DROP CONSTRAINT IF EXISTS prds_owner_id_fkey;
ALTER TABLE prds ADD CONSTRAINT prds_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Add constraint: system templates should not have organization_id
ALTER TABLE prd_templates DROP CONSTRAINT IF EXISTS chk_system_template_no_org;
ALTER TABLE prd_templates ADD CONSTRAINT chk_system_template_no_org
    CHECK ((is_system_template = FALSE) OR (organization_id IS NULL));

-- Create index for faster PRD lookups by organization
CREATE INDEX IF NOT EXISTS idx_prds_organization_id ON prds(organization_id);
CREATE INDEX IF NOT EXISTS idx_prds_owner_id ON prds(owner_id);
CREATE INDEX IF NOT EXISTS idx_prds_status ON prds(status);

-- Create index for faster version lookups
CREATE INDEX IF NOT EXISTS idx_prd_versions_prd_id ON prd_versions(prd_id);
CREATE INDEX IF NOT EXISTS idx_prd_versions_version_number ON prd_versions(prd_id, version_number);

-- Create index for organization member lookups
CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_org_id ON organization_members(organization_id);
