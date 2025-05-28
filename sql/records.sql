CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    bucket_uid TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_end TIMESTAMPTZ,
    objects_count BIGINT NOT NULL,
    bytes_total BIGINT NOT NULL
);

-- INSERT INTO records (bucket_uid, objects_count, bytes_total, period_end)
-- VALUES ('692e149b-4393-4aa8-8b54-72dfe267d202', 120, 10485760, '2025-06-30T12:00:00+00');