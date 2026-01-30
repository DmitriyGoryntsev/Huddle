CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    
    title VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Координаты в формате (lon, lat)
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    max_participants INTEGER NOT NULL DEFAULT 2,
    price NUMERIC(10, 2) DEFAULT 0,
    
    requires_approval BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'open',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Гео-индекс для быстрого поиска по радиусу
CREATE INDEX IF NOT EXISTS idx_events_location ON events USING GIST (location);