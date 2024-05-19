-- name: CreateRecode :exec
INSERT INTO recodes (season, episode, dest, origin)
VALUES (?, ?, ?, ?);

-- name: UpdatePref :exec
INSERT INTO prefs (id, rootdir)
VALUES (?, ?)
ON CONFLICT (id)
DO UPDATE SET rootdir = excluded.rootdir;

-- name: GetQueue :many
SELECT origin, dest, season, episode
FROM recodes
WHERE processed = 0;

-- name: GetPrefs :one
SELECT rootdir
FROM prefs;
