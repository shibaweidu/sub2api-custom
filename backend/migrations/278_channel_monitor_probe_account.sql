ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS probe_account_id BIGINT;

CREATE INDEX IF NOT EXISTS channel_monitors_probe_account_id
    ON channel_monitors (probe_account_id);
