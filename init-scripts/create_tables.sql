CREATE TABLE IF NOT EXISTS downloads (
    id SERIAL PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    signed_url TEXT NOT NULL,
    file_type VARCHAR(50) NOT NULL CHECK (file_type IN ('text-to-speech', 'video-to-audio', 'image-to-pdf')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
