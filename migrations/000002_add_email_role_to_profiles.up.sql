-- migrations/000002_add_email_role_to_profiles.up.sql
-- Materializa email y role en profiles (proyección read-side de auth-service).
-- Fuente de verdad sigue siendo auth_db.auth.users; este servicio mantiene una copia
-- sincronizada vía eventos USER_REGISTERED / USER_EMAIL_CHANGED / USER_ROLE_CHANGED.

ALTER TABLE "user".profiles
    ADD COLUMN IF NOT EXISTS email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS role  VARCHAR(50) NOT NULL DEFAULT 'user';

CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_email ON "user".profiles(email) WHERE email IS NOT NULL;

COMMENT ON COLUMN "user".profiles.email IS 'Proyección de auth.users.email — sincronizado por eventos. NO es la fuente de verdad.';
COMMENT ON COLUMN "user".profiles.role  IS 'Proyección de auth.users.role  — sincronizado por eventos. NO es la fuente de verdad.';
