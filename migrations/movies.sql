CREATE TABLE IF NOT EXISTS movies (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    overview TEXT,
    poster_path TEXT,
    release_date TEXT,
    language TEXT
);
