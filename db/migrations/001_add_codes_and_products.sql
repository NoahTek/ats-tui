-- Migration 001: Add candidate_code, job_code, and product tables
BEGIN;

-- 1. Add job_code to jobs if not exists
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='job_code') THEN
        ALTER TABLE jobs ADD COLUMN job_code VARCHAR(50) UNIQUE;
    END IF;
END $$;

-- 2. Add candidate_code to candidates if not exists
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='candidates' AND column_name='candidate_code') THEN
        ALTER TABLE candidates ADD COLUMN candidate_code VARCHAR(50) UNIQUE;
    END IF;
END $$;

-- 3. Create product lookup table
CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create application_products junction table
CREATE TABLE IF NOT EXISTS application_products (
    application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, product_id)
);

COMMIT;
