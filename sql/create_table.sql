CREATE TABLE videos (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    original_url TEXT,
    hls_url TEXT,
    thumbnail_url TEXT,
    duration FLOAT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    qualities JSONB DEFAULT '[]'::JSONB, -- Simpan kualitas sebagai array JSON
    hls_processed BOOLEAN DEFAULT FALSE,
    processing_error TEXT
);
