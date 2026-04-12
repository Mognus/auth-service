INSERT INTO roles (name) VALUES
    ('admin'),
    ('user'),
    ('guest')
ON CONFLICT (name) DO NOTHING;
