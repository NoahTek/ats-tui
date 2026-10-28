-- Migration 000: Initial base schema
BEGIN;

-- 1. Create custom enum type if it does not exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'application_status') THEN
        CREATE TYPE application_status AS ENUM (
            'interested',
            'follow_up',
            'call_back',
            'closed'
        );
    END IF;
END $$;

-- 2. Lookup & Reference tables
CREATE TABLE IF NOT EXISTS location (
    id SERIAL PRIMARY KEY,
    location VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS educational_qualification (
    id SERIAL PRIMARY KEY,
    qualification VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS skill (
    id SERIAL PRIMARY KEY,
    skill_name VARCHAR(100) NOT NULL,
    is_technical BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS industry (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    parent_id INTEGER REFERENCES industry(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    org_name VARCHAR(100) NOT NULL,
    org_code VARCHAR(50) UNIQUE NOT NULL,
    industry_id INTEGER REFERENCES industry(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Candidates table
CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100) UNIQUE NOT NULL,
    location_id INTEGER REFERENCES location(id) ON DELETE RESTRICT,
    qualification_id INTEGER REFERENCES educational_qualification(id) ON DELETE RESTRICT,
    dob DATE,
    resume_text TEXT,
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(resume_text, ''))
    ) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_candidates_search_vector ON candidates USING gin (search_vector);

-- 4. Jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    client_id INTEGER REFERENCES clients(id) ON DELETE RESTRICT,
    job_title VARCHAR(100) NOT NULL,
    role VARCHAR(100),
    department VARCHAR(100),
    sub_department VARCHAR(100),
    location_id INTEGER REFERENCES location(id) ON DELETE RESTRICT,
    requisition_date DATE,
    num_positions INTEGER DEFAULT 1,
    num_shared INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Applications table
CREATE TABLE IF NOT EXISTS applications (
    id SERIAL PRIMARY KEY,
    job_id INTEGER REFERENCES jobs(id) ON DELETE RESTRICT,
    candidate_id INTEGER REFERENCES candidates(id) ON DELETE CASCADE,
    skill_id INTEGER REFERENCES skill(id) ON DELETE RESTRICT,
    current_org VARCHAR(100),
    designation VARCHAR(100),
    experience_years INTEGER,
    ctc INTEGER,
    expected_ctc INTEGER,
    notice_period_in_days INTEGER,
    status application_status DEFAULT 'interested',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. Notes table
CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
    note_type VARCHAR(50),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
