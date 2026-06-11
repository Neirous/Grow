package db

const Schema = `
CREATE TABLE IF NOT EXISTS abilities (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    name              TEXT NOT NULL,
    description       TEXT NOT NULL DEFAULT '',
    base_value        REAL NOT NULL,
    current_value     REAL NOT NULL,
    growth_rate       REAL NOT NULL DEFAULT 1.0,
    decay_rate        REAL NOT NULL DEFAULT 0.5,
    last_activity_at  TEXT,
    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS activities (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS activity_effects (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id      INTEGER NOT NULL,
    ability_id       INTEGER NOT NULL,
    boost_percentage REAL NOT NULL,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (ability_id)  REFERENCES abilities(id)  ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ae_activity ON activity_effects(activity_id);
CREATE INDEX IF NOT EXISTS idx_ae_ability  ON activity_effects(ability_id);

CREATE TABLE IF NOT EXISTS activity_logs (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id  INTEGER NOT NULL,
    completed_at TEXT NOT NULL DEFAULT (datetime('now')),
    note         TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (activity_id) REFERENCES activities(id)
);

CREATE INDEX IF NOT EXISTS idx_al_activity  ON activity_logs(activity_id);
CREATE INDEX IF NOT EXISTS idx_al_completed ON activity_logs(completed_at);

CREATE TABLE IF NOT EXISTS log_ability_snapshots (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    log_id      INTEGER NOT NULL,
    ability_id  INTEGER NOT NULL,
    old_value   REAL NOT NULL,
    new_value   REAL NOT NULL,
    FOREIGN KEY (log_id)    REFERENCES activity_logs(id) ON DELETE CASCADE,
    FOREIGN KEY (ability_id) REFERENCES abilities(id)
);

CREATE INDEX IF NOT EXISTS idx_las_log     ON log_ability_snapshots(log_id);
CREATE INDEX IF NOT EXISTS idx_las_ability ON log_ability_snapshots(ability_id);

CREATE INDEX IF NOT EXISTS idx_abilities_last_activity ON abilities(last_activity_at);

CREATE TABLE IF NOT EXISTS goals (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    ability_id  INTEGER NOT NULL,
    target_value REAL NOT NULL,
    deadline    TEXT,
    is_achieved INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (ability_id) REFERENCES abilities(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`
