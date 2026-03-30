-- 001_create_core_tables.up.sql
-- Core data model for urushi-chronicle: works, process_steps, environment_readings, images

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- works: 作品テーブル
-- =============================================================================
CREATE TABLE works (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       VARCHAR(200) NOT NULL,
    description TEXT,
    technique   VARCHAR(50) NOT NULL CHECK (technique IN ('makie', 'raden', 'makie_raden', 'other')),
    material    VARCHAR(100),
    status      VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed', 'archived')),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_works_status ON works (status);
CREATE INDEX idx_works_technique ON works (technique);
CREATE INDEX idx_works_started_at ON works (started_at);

-- =============================================================================
-- process_steps: 制作工程テーブル
-- =============================================================================
CREATE TABLE process_steps (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_id     UUID NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL CHECK (char_length(name) > 0),
    description TEXT,
    step_order  INT NOT NULL,
    category    VARCHAR(50) NOT NULL CHECK (category IN (
        'shitanuri',   -- 下塗り
        'nakanuri',    -- 中塗り
        'uwanuri',     -- 上塗り
        'makie',       -- 蒔絵
        'raden',       -- 螺鈿
        'togidashi',   -- 研ぎ出し
        'roiro',       -- 呂色仕上げ
        'other'        -- その他
    )),
    materials_used JSONB DEFAULT '[]'::jsonb,
    notes       TEXT,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (work_id, step_order)
);

CREATE INDEX idx_process_steps_work_id ON process_steps (work_id);
CREATE INDEX idx_process_steps_category ON process_steps (category);

-- =============================================================================
-- environment_readings: 環境データテーブル（TimescaleDB hypertable）
-- =============================================================================
CREATE TABLE environment_readings (
    time          TIMESTAMPTZ NOT NULL,
    sensor_id     VARCHAR(50) NOT NULL,
    location      VARCHAR(100) NOT NULL DEFAULT 'urushi_buro',
    temperature   DOUBLE PRECISION NOT NULL CHECK (temperature >= -10.0 AND temperature <= 100.0),
    humidity      DOUBLE PRECISION NOT NULL CHECK (humidity >= 0.0 AND humidity <= 100.0),
    work_id       UUID REFERENCES works(id) ON DELETE SET NULL,
    process_step_id UUID REFERENCES process_steps(id) ON DELETE SET NULL
);

-- Convert to TimescaleDB hypertable for efficient time-series queries
SELECT create_hypertable('environment_readings', 'time');

CREATE INDEX idx_environment_readings_sensor ON environment_readings (sensor_id, time DESC);
CREATE INDEX idx_environment_readings_work ON environment_readings (work_id, time DESC);
CREATE INDEX idx_environment_readings_step ON environment_readings (process_step_id, time DESC);

-- =============================================================================
-- images: 画像テーブル
-- =============================================================================
CREATE TABLE images (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_id         UUID NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    process_step_id UUID REFERENCES process_steps(id) ON DELETE SET NULL,
    file_path       TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL CHECK (file_size_bytes > 0 AND file_size_bytes <= 10485760), -- max 10MB
    content_type    VARCHAR(20) NOT NULL CHECK (content_type IN ('image/jpeg', 'image/png')),
    image_type      VARCHAR(20) NOT NULL DEFAULT 'process' CHECK (image_type IN ('process', 'macro', 'aging', 'overview')),
    caption         VARCHAR(500),
    taken_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_images_work_id ON images (work_id);
CREATE INDEX idx_images_process_step_id ON images (process_step_id);
CREATE INDEX idx_images_image_type ON images (image_type);

-- =============================================================================
-- alert_thresholds: アラート閾値設定テーブル
-- =============================================================================
CREATE TABLE alert_thresholds (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sensor_id       VARCHAR(50) NOT NULL,
    temperature_min DOUBLE PRECISION NOT NULL DEFAULT 20.0 CHECK (temperature_min >= -10.0 AND temperature_min <= 100.0),
    temperature_max DOUBLE PRECISION NOT NULL DEFAULT 30.0 CHECK (temperature_max >= -10.0 AND temperature_max <= 100.0),
    humidity_min    DOUBLE PRECISION NOT NULL DEFAULT 70.0 CHECK (humidity_min >= 0.0 AND humidity_min <= 100.0),
    humidity_max    DOUBLE PRECISION NOT NULL DEFAULT 85.0 CHECK (humidity_max >= 0.0 AND humidity_max <= 100.0),
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (temperature_min < temperature_max),
    CHECK (humidity_min < humidity_max)
);

CREATE INDEX idx_alert_thresholds_sensor ON alert_thresholds (sensor_id);

-- =============================================================================
-- updated_at trigger function
-- =============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_works_updated_at
    BEFORE UPDATE ON works
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_process_steps_updated_at
    BEFORE UPDATE ON process_steps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_alert_thresholds_updated_at
    BEFORE UPDATE ON alert_thresholds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
