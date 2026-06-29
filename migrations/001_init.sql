CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  app_name TEXT NOT NULL,
  event_type TEXT NOT NULL CHECK (event_type IN ('open', 'close')),
  timestamp TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_user_timestamp ON events(user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_user_app_timestamp ON events(user_id, app_name, timestamp);

CREATE TABLE IF NOT EXISTS alerts (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  type TEXT NOT NULL,
  severity TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_created_at ON alerts(user_id, created_at);

CREATE TABLE IF NOT EXISTS rules (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  app_name TEXT NOT NULL,
  max_opens INTEGER NOT NULL,
  window_minutes INTEGER NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE
);
