-- name: CreateStreak :one
INSERT INTO streaks (user_id, current_streak, longest_streak, completed_today, last_completed_at)
VALUES ($1, 0, 0, FALSE, NULL)
RETURNING id, user_id, current_streak, longest_streak, completed_today, last_completed_at, created_at, updated_at;

-- name: GetStreakByUserID :one
SELECT id, user_id, current_streak, longest_streak, completed_today, last_completed_at, created_at, updated_at
FROM streaks
WHERE user_id = $1;

-- name: MarkStreakCompletedToday :one
UPDATE streaks
SET completed_today = TRUE,
    last_completed_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1
RETURNING id, user_id, current_streak, longest_streak, completed_today, last_completed_at, created_at, updated_at;

-- name: ResetDailyCompletionFlags :exec
UPDATE streaks
SET completed_today = FALSE,
    updated_at = NOW()
WHERE completed_today = TRUE;
