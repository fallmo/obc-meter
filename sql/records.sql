CREATE TABLE runs (
    id SERIAL PRIMARY KEY,
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    trigger TEXT NOT NULL,
    all_uids TEXT[] NOT NULL,
    failed_uids TEXT[] NOT NULL,
    error_messages TEXT[] NOT NULL
);

CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    bucket_uid TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_end TIMESTAMPTZ,
    objects_count BIGINT NOT NULL,
    bytes_total BIGINT NOT NULL,
    run_id INT NOT NULL REFERENCES runs(id)
);

-- INSERT INTO records (bucket_uid, objects_count, bytes_total, period_end)
-- VALUES ('692e149b-4393-4aa8-8b54-72dfe267d202', 120, 10485760, '2025-06-30T12:00:00+00');