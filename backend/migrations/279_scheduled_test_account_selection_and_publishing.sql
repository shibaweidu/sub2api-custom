ALTER TABLE scheduled_test_plans
    ADD COLUMN IF NOT EXISTS account_ids BIGINT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS auto_publish_results BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE scheduled_test_results
    ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_scheduled_test_results_unpublished
    ON scheduled_test_results (plan_id, started_at DESC)
    WHERE published = FALSE;
