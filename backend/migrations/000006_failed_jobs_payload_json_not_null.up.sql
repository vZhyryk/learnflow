-- No backfill needed: every INSERT INTO failed_jobs in application code
-- (internal/worker/worker.go, dlq_retry.go) always sets payload_json.
ALTER TABLE failed_jobs
    ALTER COLUMN payload_json SET NOT NULL;