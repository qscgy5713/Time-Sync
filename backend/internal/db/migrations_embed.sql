CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    timezone VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS event_options (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    start_datetime TIMESTAMPTZ NOT NULL,
    end_datetime TIMESTAMPTZ NOT NULL
);

-- Nullable: once set, the event is locked to this slot and voting closes.
ALTER TABLE events ADD COLUMN IF NOT EXISTS finalized_option_id BIGINT REFERENCES event_options(id);

-- Secret held only by the organizer's browser; required to finalize/unfinalize.
-- NULL for events created before this existed (they can never be finalized).
ALTER TABLE events ADD COLUMN IF NOT EXISTS owner_token UUID;

CREATE INDEX IF NOT EXISTS idx_event_options_event_id ON event_options(event_id);

CREATE TABLE IF NOT EXISTS participants (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    edit_token UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE participants ADD COLUMN IF NOT EXISTS edit_token UUID;

CREATE INDEX IF NOT EXISTS idx_participants_event_id ON participants(event_id);

CREATE TABLE IF NOT EXISTS availabilities (
    participant_id BIGINT NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    event_option_id BIGINT NOT NULL REFERENCES event_options(id) ON DELETE CASCADE,
    PRIMARY KEY (participant_id, event_option_id)
);
