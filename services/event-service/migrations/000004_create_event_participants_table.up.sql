CREATE TABLE IF NOT EXISTS event_participants (
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (event_id, user_id)
);

-- Индекс для быстрого получения всех участников конкретного ивента
CREATE INDEX IF NOT EXISTS idx_event_participants_event_id ON event_participants(event_id);