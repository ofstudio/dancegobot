CREATE TABLE "events"
(
    "id"         TEXT PRIMARY KEY,
    "owner_id"   INTEGER   NOT NULL,
    "data"       JSON      NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE "users"
(
    "id"         INTEGER PRIMARY KEY,
    "profile"    JSON      NOT NULL,
    "session"    JSON      NOT NULL DEFAULT ('{}'),
    "settings"   JSON      NOT NULL DEFAULT ('{}'),
    "created_at" TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    "updated_at" TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE "history"
(
    id         INTEGER PRIMARY KEY,
    action     TEXT      NOT NULL,
    profile_id INTEGER,
    event_id   TEXT,
    data       JSON      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
