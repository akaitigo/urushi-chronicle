-- 001_create_core_tables.down.sql
-- Rollback core data model tables

DROP TRIGGER IF EXISTS trigger_alert_thresholds_updated_at ON alert_thresholds;
DROP TRIGGER IF EXISTS trigger_process_steps_updated_at ON process_steps;
DROP TRIGGER IF EXISTS trigger_works_updated_at ON works;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS alert_thresholds;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS environment_readings;
DROP TABLE IF EXISTS process_steps;
DROP TABLE IF EXISTS works;

DROP EXTENSION IF EXISTS timescaledb;
DROP EXTENSION IF EXISTS "uuid-ossp";
