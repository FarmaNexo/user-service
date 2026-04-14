-- migrations/000002_add_email_role_to_profiles.down.sql
DROP INDEX IF EXISTS "user".idx_profiles_email;
ALTER TABLE "user".profiles
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS role;
