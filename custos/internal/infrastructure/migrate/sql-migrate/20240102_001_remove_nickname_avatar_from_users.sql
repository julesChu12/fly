-- +migrate Up
-- Migrate existing data from users to user_profiles
INSERT INTO user_profiles (user_id, nickname, avatar, gender, created_at, updated_at)
SELECT
    id,
    COALESCE(nickname, ''),
    COALESCE(avatar, ''),
    'other',
    created_at,
    NOW()
FROM users
WHERE id NOT IN (SELECT user_id FROM user_profiles)
ON DUPLICATE KEY UPDATE
    nickname = COALESCE(VALUES(nickname), user_profiles.nickname),
    avatar = COALESCE(VALUES(avatar), user_profiles.avatar),
    updated_at = NOW();

-- Update existing profiles with users data if users data is not empty
UPDATE user_profiles up
INNER JOIN users u ON up.user_id = u.id
SET
    up.nickname = CASE
        WHEN u.nickname IS NOT NULL AND u.nickname != '' AND (up.nickname IS NULL OR up.nickname = '')
        THEN u.nickname
        ELSE up.nickname
    END,
    up.avatar = CASE
        WHEN u.avatar IS NOT NULL AND u.avatar != '' AND (up.avatar IS NULL OR up.avatar = '')
        THEN u.avatar
        ELSE up.avatar
    END,
    up.updated_at = NOW()
WHERE (u.nickname IS NOT NULL AND u.nickname != '' AND (up.nickname IS NULL OR up.nickname = ''))
   OR (u.avatar IS NOT NULL AND u.avatar != '' AND (up.avatar IS NULL OR up.avatar = ''));

-- Remove nickname and avatar columns from users table
ALTER TABLE users DROP COLUMN nickname;
ALTER TABLE users DROP COLUMN avatar;

-- +migrate Down
-- Restore nickname and avatar columns
ALTER TABLE users
    ADD COLUMN nickname VARCHAR(100) AFTER password,
    ADD COLUMN avatar VARCHAR(255) AFTER nickname;

-- Migrate data back from user_profiles to users
UPDATE users u
INNER JOIN user_profiles up ON u.id = up.user_id
SET
    u.nickname = up.nickname,
    u.avatar = up.avatar;
