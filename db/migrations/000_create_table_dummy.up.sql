CREATE TABLE IF NOT EXISTS dummy (
  user_id       UUID PRIMARY KEY,
  username     TEXT NOT NULL,
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ
);
