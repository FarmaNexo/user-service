-- migrations/000001_init_schema.down.sql
-- Rollback de la migración inicial

-- Eliminar tablas (en orden)
DROP TABLE IF EXISTS "user".preferences;
DROP TABLE IF EXISTS "user".addresses;
DROP TABLE IF EXISTS "user".profiles;

-- Eliminar esquema
DROP SCHEMA IF EXISTS "user";
