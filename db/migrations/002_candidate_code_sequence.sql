-- Migration 002: Candidate code sequence and CAN-xxx convention
BEGIN;

-- 1. Create sequence for candidate codes if not exists
CREATE SEQUENCE IF NOT EXISTS candidate_code_seq START WITH 11;

-- 2. Update existing candidates with CAN-001 through CAN-010 where code is not set
UPDATE candidates SET candidate_code = 'CAN-' || LPAD(id::text, 3, '0') WHERE candidate_code IS NULL;

-- 3. Set default value for candidate_code column using the sequence
ALTER TABLE candidates ALTER COLUMN candidate_code SET DEFAULT ('CAN-' || LPAD(nextval('candidate_code_seq')::text, 3, '0'));

COMMIT;
