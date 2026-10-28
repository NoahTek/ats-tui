-- Comprehensive seed script for ATS database
BEGIN;

-- Clean existing data for clean reseed
TRUNCATE TABLE notes, application_products, applications, candidates, jobs, clients, product, educational_qualification, location, skill, industry RESTART IDENTITY CASCADE;

-- 1. Locations
INSERT INTO location (location) VALUES 
    ('San Francisco, CA'),
    ('New York, NY'),
    ('Austin, TX'),
    ('Bengaluru, India'),
    ('Remote')
ON CONFLICT (location) DO NOTHING;

-- 2. Educational Qualifications
INSERT INTO educational_qualification (qualification) VALUES 
    ('B.S. in Computer Science'),
    ('M.S. in Software Engineering'),
    ('B.Tech in Information Technology'),
    ('B.A. in Mathematics'),
    ('Self-Taught / Bootcamp')
ON CONFLICT (qualification) DO NOTHING;

-- 3. Skills
INSERT INTO skill (skill_name, is_technical) VALUES
    ('Go', true),
    ('PostgreSQL', true),
    ('Kubernetes', true),
    ('Docker', true),
    ('Distributed Systems', true),
    ('TypeScript', true),
    ('System Architecture', true),
    ('Technical Leadership', false)
ON CONFLICT DO NOTHING;

-- 4. Industry & Sub-categories
-- Top-level industries
INSERT INTO industry (id, name, parent_id) VALUES 
    (1, 'Information Technology', NULL),
    (2, 'Financial Services', NULL),
    (3, 'Developer Platforms', NULL);

-- Sub-categories
INSERT INTO industry (id, name, parent_id) VALUES 
    (10, 'Cloud Infrastructure', 1),
    (11, 'Developer Tools', 1),
    (20, 'Payment Systems', 2),
    (21, 'Quantitative Trading', 2),
    (30, 'API Platform', 3);

-- 5. Products
INSERT INTO product (name) VALUES
    ('Payment Gateway Engine'),
    ('Fraud Risk Scoring'),
    ('Kubernetes Custom Controller'),
    ('High-Frequency Order Book'),
    ('Cloud SRE Observability Suite'),
    ('Distributed Cache');

-- 6. Clients (Organisations)
INSERT INTO clients (org_name, org_code, industry_id) VALUES
    ('Stripe-like Fintech Inc.', 'ABCL', 20),
    ('CloudNative Labs', 'YSIL', 10),
    ('Apex Quant Systems', 'APEX', 21),
    ('ScaleDev Systems', 'SDEV', 11)
ON CONFLICT (org_code) DO NOTHING;

-- 7. Jobs (Open Positions)
INSERT INTO jobs (job_code, client_id, job_title, role, department, sub_department, location_id, requisition_date, num_positions, num_shared) VALUES
    (
        'JOB1',
        (SELECT id FROM clients WHERE org_code = 'YSIL'),
        'Senior Go Backend Engineer',
        'Senior Software Engineer',
        'Engineering',
        'Infrastructure Platform',
        (SELECT id FROM location WHERE location = 'Remote'),
        '2026-01-15',
        3,
        5
    ),
    (
        'JOB2',
        (SELECT id FROM clients WHERE org_code = 'ABCL'),
        'Staff Systems Engineer',
        'Staff Engineer',
        'Core Banking',
        'Transaction Processing',
        (SELECT id FROM location WHERE location = 'San Francisco, CA'),
        '2026-02-01',
        2,
        3
    ),
    (
        'JOB3',
        (SELECT id FROM clients WHERE org_code = 'APEX'),
        'High-Frequency Trading Engineer',
        'Senior Systems Developer',
        'Trading Tech',
        'Low-Latency Execution',
        (SELECT id FROM location WHERE location = 'New York, NY'),
        '2026-02-10',
        1,
        2
    );

-- 8. Candidates (Master Candidate Pool)
INSERT INTO candidates (candidate_code, name, email, phone, location_id, qualification_id, dob, resume_text) VALUES
    (
        'CAN-001',
        'Alex Rivera',
        'alex.rivera@example.com',
        '+1-555-0101',
        (SELECT id FROM location WHERE location = 'San Francisco, CA'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.S. in Computer Science'),
        '1994-05-12',
        'Experienced Go and distributed systems developer. Built high-throughput microservices using gRPC, pgx, and Kafka. Skilled in Docker, Kubernetes, and PostgreSQL query optimization.'
    ),
    (
        'CAN-002',
        'Priya Sharma',
        'priya.sharma@example.com',
        '+91-98765-43210',
        (SELECT id FROM location WHERE location = 'Bengaluru, India'),
        (SELECT id FROM educational_qualification WHERE qualification = 'M.S. in Software Engineering'),
        '1992-09-24',
        'Backend engineer with 6 years building payment gateways in Go and Java. Deep familiarity with PostgreSQL concurrency, ACID transactions, and Redis caching.'
    ),
    (
        'CAN-003',
        'Marcus Vance',
        'marcus.vance@example.com',
        '+1-555-0144',
        (SELECT id FROM location WHERE location = 'Austin, TX'),
        (SELECT id FROM educational_qualification WHERE qualification = 'Self-Taught / Bootcamp'),
        '1996-11-03',
        'Self-taught systems hacker. Heavy contributor to open source CLI tools written in Go and Rust. Proficient in Linux kernel basics, eBPF, and TUI development.'
    ),
    (
        'CAN-004',
        'Elena Rostova',
        'elena.rostova@example.com',
        '+1-555-0189',
        (SELECT id FROM location WHERE location = 'New York, NY'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.A. in Mathematics'),
        '1993-03-18',
        'Quantitative backend developer specializing in financial order matching engines. Proficient in Go concurrency primitives, memory management, and SQL tuning.'
    ),
    (
        'CAN-005',
        'Devon Miller',
        'devon.miller@example.com',
        '+1-555-0199',
        (SELECT id FROM location WHERE location = 'Remote'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.Tech in Information Technology'),
        '1991-07-30',
        'DevOps & Platform engineer. Expert in Kubernetes operators written in Go, Helm charts, and automated multi-region PostgreSQL disaster recovery.'
    ),
    (
        'CAN-006',
        'Aisha Khan',
        'aisha.khan@example.com',
        '+1-555-0211',
        (SELECT id FROM location WHERE location = 'Remote'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.S. in Computer Science'),
        '1989-12-05',
        'Cloud infrastructure architect with 8 years of experience. Designed resilient distributed architectures on AWS and GCP using Go, Kubernetes, and Terraform.'
    ),
    (
        'CAN-007',
        'Lucas Silva',
        'lucas.silva@example.com',
        '+1-555-0233',
        (SELECT id FROM location WHERE location = 'San Francisco, CA'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.S. in Computer Science'),
        '1995-04-19',
        'Full-stack engineer leaning heavily towards backend Go and database internals. Created real-time collaborative applications with WebSockets.'
    ),
    (
        'CAN-008',
        'Hannah Schmidt',
        'hannah.schmidt@example.com',
        '+1-555-0255',
        (SELECT id FROM location WHERE location = 'New York, NY'),
        (SELECT id FROM educational_qualification WHERE qualification = 'M.S. in Software Engineering'),
        '1990-08-14',
        'Site reliability engineer specializing in zero-downtime database migrations, pgx connection pooling, and Prometheus observability.'
    ),
    (
        'CAN-009',
        'Chen Wei',
        'chen.wei@example.com',
        '+1-555-0277',
        (SELECT id FROM location WHERE location = 'Austin, TX'),
        (SELECT id FROM educational_qualification WHERE qualification = 'B.S. in Computer Science'),
        '1993-01-22',
        'Distributed systems engineer. Worked on consensus algorithms (Raft), key-value stores in Go, and high-performance network programming.'
    ),
    (
        'CAN-010',
        'Jordan Bell',
        'jordan.bell@example.com',
        '+1-555-0299',
        (SELECT id FROM location WHERE location = 'Remote'),
        (SELECT id FROM educational_qualification WHERE qualification = 'Self-Taught / Bootcamp'),
        '1994-10-09',
        'Product-minded backend engineer with experience across Go, Node.js, and Postgres. Great communicator with startup leadership background.'
    )
ON CONFLICT (email) DO NOTHING;

-- 9. Applications (Submissions)
INSERT INTO applications (job_id, candidate_id, skill_id, current_org, designation, experience_years, ctc, expected_ctc, notice_period_in_days, status) VALUES
    -- Interested
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB1'),
        (SELECT id FROM candidates WHERE email = 'alex.rivera@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Go' LIMIT 1),
        'DataCore Tech',
        'Software Engineer II',
        4,
        135000,
        155000,
        30,
        'interested'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB1'),
        (SELECT id FROM candidates WHERE email = 'marcus.vance@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Go' LIMIT 1),
        'OpenSource Studio',
        'Senior CLI Engineer',
        5,
        140000,
        160000,
        15,
        'interested'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB2'),
        (SELECT id FROM candidates WHERE email = 'lucas.silva@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Distributed Systems' LIMIT 1),
        'SyncWave',
        'Backend Developer',
        3,
        120000,
        140000,
        30,
        'interested'
    ),

    -- Follow-Up
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB2'),
        (SELECT id FROM candidates WHERE email = 'priya.sharma@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'PostgreSQL' LIMIT 1),
        'PayGlobal Systems',
        'Lead Backend Engineer',
        6,
        150000,
        175000,
        60,
        'follow_up'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB3'),
        (SELECT id FROM candidates WHERE email = 'elena.rostova@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Go' LIMIT 1),
        'Apex Quant Labs',
        'Systems Software Engineer',
        5,
        160000,
        185000,
        30,
        'follow_up'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB3'),
        (SELECT id FROM candidates WHERE email = 'chen.wei@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Distributed Systems' LIMIT 1),
        'KVS Networks',
        'Staff Infrastructure Engineer',
        7,
        175000,
        195000,
        45,
        'follow_up'
    ),

    -- Call-Back
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB1'),
        (SELECT id FROM candidates WHERE email = 'devon.miller@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Kubernetes' LIMIT 1),
        'InfraCloud Inc.',
        'Senior DevOps Specialist',
        6,
        155000,
        170000,
        15,
        'call_back'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB1'),
        (SELECT id FROM candidates WHERE email = 'hannah.schmidt@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'PostgreSQL' LIMIT 1),
        'ReliableScale SRE',
        'Senior Database Reliability Engineer',
        6,
        160000,
        180000,
        30,
        'call_back'
    ),

    -- Closed
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB2'),
        (SELECT id FROM candidates WHERE email = 'aisha.khan@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'System Architecture' LIMIT 1),
        'NextGen Cloud Co.',
        'Principal Architect',
        8,
        190000,
        210000,
        30,
        'closed'
    ),
    (
        (SELECT id FROM jobs WHERE job_code = 'JOB1'),
        (SELECT id FROM candidates WHERE email = 'jordan.bell@example.com'),
        (SELECT id FROM skill WHERE skill_name = 'Technical Leadership' LIMIT 1),
        'RapidLaunch Ventures',
        'Tech Lead',
        5,
        145000,
        155000,
        15,
        'closed'
    );

-- 10. Link Products to Applications
INSERT INTO application_products (application_id, product_id)
SELECT a.id, p.id FROM applications a, product p 
WHERE a.candidate_id = (SELECT id FROM candidates WHERE email = 'alex.rivera@example.com') 
  AND p.name IN ('Distributed Cache', 'Cloud SRE Observability Suite');

INSERT INTO application_products (application_id, product_id)
SELECT a.id, p.id FROM applications a, product p 
WHERE a.candidate_id = (SELECT id FROM candidates WHERE email = 'priya.sharma@example.com') 
  AND p.name IN ('Payment Gateway Engine', 'Fraud Risk Scoring');

INSERT INTO application_products (application_id, product_id)
SELECT a.id, p.id FROM applications a, product p 
WHERE a.candidate_id = (SELECT id FROM candidates WHERE email = 'devon.miller@example.com') 
  AND p.name IN ('Kubernetes Custom Controller', 'Cloud SRE Observability Suite');

INSERT INTO application_products (application_id, product_id)
SELECT a.id, p.id FROM applications a, product p 
WHERE a.candidate_id = (SELECT id FROM candidates WHERE email = 'elena.rostova@example.com') 
  AND p.name IN ('High-Frequency Order Book');

-- 11. Notes (Application Comments / History)
INSERT INTO notes (application_id, note_type, content, created_at) VALUES
    (
        (SELECT a.id FROM applications a JOIN candidates c ON a.candidate_id = c.id WHERE c.email = 'alex.rivera@example.com' LIMIT 1),
        'Comment',
        'Profile shared with hiring manager. Initial review passed; candidate possesses strong Go and gRPC foundations.',
        '2026-02-12 10:30:00+00'
    ),
    (
        (SELECT a.id FROM applications a JOIN candidates c ON a.candidate_id = c.id WHERE c.email = 'priya.sharma@example.com' LIMIT 1),
        'Comment',
        'Initial technical screening completed. Extensive payment systems experience; requested follow-up interview on concurrency.',
        '2026-02-14 14:00:00+00'
    ),
    (
        (SELECT a.id FROM applications a JOIN candidates c ON a.candidate_id = c.id WHERE c.email = 'priya.sharma@example.com' LIMIT 1),
        'Feedback',
        'Follow-up call with engineering director went exceptionally well. Candidate is aligned on comp and 60-day notice period.',
        '2026-02-18 16:30:00+00'
    ),
    (
        (SELECT a.id FROM applications a JOIN candidates c ON a.candidate_id = c.id WHERE c.email = 'devon.miller@example.com' LIMIT 1),
        'Feedback',
        'Passed technical deep-dive on Kubernetes operator reconciliation loop. Scheduling final executive call-back.',
        '2026-02-19 11:15:00+00'
    ),
    (
        (SELECT a.id FROM applications a JOIN candidates c ON a.candidate_id = c.id WHERE c.email = 'aisha.khan@example.com' LIMIT 1),
        'Comment',
        'Offer accepted! Joining next month as Principal Cloud Architect.',
        '2026-02-25 09:00:00+00'
    );

COMMIT;
