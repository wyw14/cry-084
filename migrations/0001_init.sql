CREATE TABLE IF NOT EXISTS assets (
    id text PRIMARY KEY,
    mall_id text NOT NULL,
    code text NOT NULL,
    version bigint NOT NULL CHECK (version > 0),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (mall_id, code)
);
CREATE TABLE IF NOT EXISTS asset_events (
    id text PRIMARY KEY,
    asset_id text NOT NULL REFERENCES assets(id),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS asset_events_timeline_idx ON asset_events(asset_id, created_at, id);
CREATE TABLE IF NOT EXISTS routes (id text PRIMARY KEY, mall_id text NOT NULL, version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL);
CREATE TABLE IF NOT EXISTS schedules (id text PRIMARY KEY, route_id text NOT NULL REFERENCES routes(id), payload jsonb NOT NULL);
CREATE TABLE IF NOT EXISTS template_versions (id text PRIMARY KEY, template_id text NOT NULL, version_number integer NOT NULL CHECK(version_number>0), payload jsonb NOT NULL, UNIQUE(template_id,version_number));
CREATE TABLE IF NOT EXISTS inspection_tasks (id text PRIMARY KEY, route_id text NOT NULL REFERENCES routes(id), version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS inspection_results (id text PRIMARY KEY, task_id text NOT NULL REFERENCES inspection_tasks(id), asset_id text NOT NULL REFERENCES assets(id), offline_key text NOT NULL UNIQUE, payload jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS hazards (id text PRIMARY KEY, asset_id text NOT NULL REFERENCES assets(id), version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS parts (id text PRIMARY KEY, mall_id text NOT NULL, sku text NOT NULL, version bigint NOT NULL CHECK(version>0), payload jsonb NOT NULL, UNIQUE(mall_id,sku));
CREATE TABLE IF NOT EXISTS part_consumptions (id text PRIMARY KEY, payload jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS audit_events (id text PRIMARY KEY, payload jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS idempotency_keys (scope text NOT NULL, key text NOT NULL, request_hash text NOT NULL, status integer NOT NULL, response bytea NOT NULL, expires_at timestamptz NOT NULL, PRIMARY KEY(scope,key));
CREATE TABLE IF NOT EXISTS outbox (id text PRIMARY KEY, topic text NOT NULL, payload jsonb NOT NULL, attempts integer NOT NULL DEFAULT 0, available_at timestamptz NOT NULL, completed_at timestamptz, last_error text NOT NULL DEFAULT '');
CREATE INDEX IF NOT EXISTS outbox_claim_idx ON outbox(available_at) WHERE completed_at IS NULL;
