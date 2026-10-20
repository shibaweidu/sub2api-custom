ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS quality_group_id BIGINT;

CREATE INDEX IF NOT EXISTS channelmonitor_quality_group_id
    ON channel_monitors (quality_group_id);
