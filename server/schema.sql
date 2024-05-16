CREATE TABLE IF NOT EXISTS recodes (
    id INT PRIMARY KEY,
    origin TEXT NOT NULL,
    dest TEXT NOT NULL,
    season TEXT NOT NULL,
    episode TEXT NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT(0),
    createdAt INTEGER NOT NULL DEFAULT(unixepoch(CURRENT_TIMESTAMP)),
    updatedAt INTEGER NOT NULL DEFAULT(unixepoch(CURRENT_TIMESTAMP))
);

CREATE TABLE IF NOT EXISTS prefs (
    id INT PRIMARY KEY CHECK (id = 1),
    rootdir TEXT
);
