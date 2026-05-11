-- Phase 1 migration: bring all previously-Mongo collections into Postgres.
-- Tables are designed for direct mapping from existing models (see internal/model/*.go).
-- Indexed/queried fields are columns; flexible payloads use JSONB.

-- Shared trigger function: keeps updated_at fresh on UPDATE.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- bond_issuers  (was: MongoDB bond_issuers)
-- Maps SECID → emitter. `inn` and `external_ids` are reserved for Phase 5 (Tinkoff).
-- ============================================================================
CREATE TABLE IF NOT EXISTS bond_issuers (
    secid        VARCHAR(64)  PRIMARY KEY,
    emitter_id   BIGINT       NOT NULL DEFAULT 0,
    emitter_name TEXT         NOT NULL DEFAULT '',
    inn          TEXT,
    external_ids JSONB,
    is_hidden    BOOLEAN      NOT NULL DEFAULT FALSE,
    needs_sync   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_bond_issuers_emitter_id ON bond_issuers (emitter_id);

DROP TRIGGER IF EXISTS trg_bond_issuers_updated ON bond_issuers;
CREATE TRIGGER trg_bond_issuers_updated BEFORE UPDATE ON bond_issuers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- issuer_ratings  (was: MongoDB issuer_ratings)
-- Composite PK (emitter_id, agency) — matches existing Mongo unique index.
-- ============================================================================
CREATE TABLE IF NOT EXISTS issuer_ratings (
    emitter_id  BIGINT       NOT NULL,
    agency      VARCHAR(64)  NOT NULL,
    issuer      TEXT         NOT NULL DEFAULT '',
    rating      VARCHAR(128) NOT NULL,
    score       INTEGER      NOT NULL DEFAULT 0,
    score_ord   INTEGER      NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (emitter_id, agency)
);
CREATE INDEX IF NOT EXISTS idx_issuer_ratings_emitter_id ON issuer_ratings (emitter_id);

DROP TRIGGER IF EXISTS trg_issuer_ratings_updated ON issuer_ratings;
CREATE TRIGGER trg_issuer_ratings_updated BEFORE UPDATE ON issuer_ratings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- dohod_details  (was: MongoDB dohod_details, TTL 30d)
-- ISIN is the natural key. ~80 dohod fields packed into JSONB `data` —
-- никто из бэка не запрашивает их по value, всё читается целиком.
-- Cleanup of stale rows: nightly DELETE WHERE updated_at < now() - interval '30 days'.
-- ============================================================================
CREATE TABLE IF NOT EXISTS dohod_details (
    isin         VARCHAR(64)  PRIMARY KEY,
    secid        VARCHAR(64),
    issuer_name  TEXT         NOT NULL DEFAULT '',
    data         JSONB        NOT NULL,
    fetched_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_dohod_details_secid ON dohod_details (secid);
CREATE INDEX IF NOT EXISTS idx_dohod_details_updated ON dohod_details (updated_at);

DROP TRIGGER IF EXISTS trg_dohod_details_updated ON dohod_details;
CREATE TRIGGER trg_dohod_details_updated BEFORE UPDATE ON dohod_details
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- bond_analyses  (was: MongoDB bond_analyses)
-- Indexes для: запросов по secid отсортированных по дате; bulk-stats аггрегаты.
-- ID — UUID-строка (совместимо с прежним Mongo ObjectId-as-string в API).
-- ============================================================================
CREATE TABLE IF NOT EXISTS bond_analyses (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    secid       VARCHAR(64)  NOT NULL,
    response    TEXT         NOT NULL,
    rating      DOUBLE PRECISION,
    json_data   JSONB,
    custom_json JSONB,
    user_id     BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    timestamp   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    saved_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    tags        TEXT[]
);
CREATE INDEX IF NOT EXISTS idx_bond_analyses_secid_ts ON bond_analyses (secid, timestamp DESC);

-- ============================================================================
-- queue_jobs  (was: MongoDB queue_jobs)
-- Dedup index — partial index on pending/running for fast lookup.
-- ============================================================================
CREATE TABLE IF NOT EXISTS queue_jobs (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    type         VARCHAR(64)  NOT NULL,
    secid        VARCHAR(64),
    reference_id TEXT,
    status       VARCHAR(32)  NOT NULL,
    data         JSONB,
    result       JSONB,
    error        TEXT         NOT NULL DEFAULT '',
    attempts     INTEGER      NOT NULL DEFAULT 0,
    max_attempts INTEGER      NOT NULL DEFAULT 3,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_dedup ON queue_jobs (type, secid, status)
    WHERE status IN ('pending', 'running');
CREATE INDEX IF NOT EXISTS idx_queue_jobs_pending ON queue_jobs (status, created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_queue_jobs_stale ON queue_jobs (status, updated_at)
    WHERE status = 'running';

DROP TRIGGER IF EXISTS trg_queue_jobs_updated ON queue_jobs;
CREATE TRIGGER trg_queue_jobs_updated BEFORE UPDATE ON queue_jobs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- chat_sessions / chat_messages  (was: MongoDB chat_sessions / chat_messages)
-- ============================================================================
CREATE TABLE IF NOT EXISTS chat_sessions (
    session_id  VARCHAR(64)  PRIMARY KEY,
    title       TEXT         NOT NULL DEFAULT '',
    agent_type  VARCHAR(64)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id          BIGSERIAL    PRIMARY KEY,
    session_id  VARCHAR(64)  NOT NULL REFERENCES chat_sessions(session_id) ON DELETE CASCADE,
    role        VARCHAR(32)  NOT NULL,
    content     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session_created ON chat_messages (session_id, created_at);

-- ============================================================================
-- Phase 2 — scoring engine (reserved tables)
-- ============================================================================
CREATE TABLE IF NOT EXISTS scoring_profiles (
    code        VARCHAR(64)  PRIMARY KEY,
    name        TEXT         NOT NULL,
    is_preset   BOOLEAN      NOT NULL DEFAULT FALSE,
    user_id     BIGINT       REFERENCES users(id) ON DELETE CASCADE,
    weights     JSONB        NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_scoring_profiles_updated ON scoring_profiles;
CREATE TRIGGER trg_scoring_profiles_updated BEFORE UPDATE ON scoring_profiles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Bond scores keep history; query "latest" via ORDER BY computed_at DESC LIMIT 1.
CREATE TABLE IF NOT EXISTS bond_scores (
    id           BIGSERIAL    PRIMARY KEY,
    secid        VARCHAR(64)  NOT NULL,
    profile_code VARCHAR(64)  NOT NULL,
    score        REAL         NOT NULL,
    breakdown    JSONB        NOT NULL,
    computed_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_bond_scores_secid_profile_ts
    ON bond_scores (secid, profile_code, computed_at DESC);

CREATE TABLE IF NOT EXISTS bond_score_explanations (
    id            BIGSERIAL    PRIMARY KEY,
    bond_score_id BIGINT       NOT NULL REFERENCES bond_scores(id) ON DELETE CASCADE,
    llm_model     VARCHAR(128) NOT NULL,
    text          TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_bond_score_explanations_score
    ON bond_score_explanations (bond_score_id);

-- Seed three preset profiles. Weights mirror docs/roadmap.md Phase 2.
INSERT INTO scoring_profiles (code, name, is_preset, weights) VALUES
    ('low',  'Низкий риск',      TRUE,
        '{"credit_rating":0.40,"ytm":0.05,"ytm_premium":0.05,"duration":0.15,"liquidity":0.15,"category":0.05,"put_offer_soon":-0.05,"issue_size":0.05,"coupon_type":0.05,"rating_age":0.05,"dohod_quality":0.00,"dohod_stability":0.00}'::JSONB),
    ('mid',  'Средний риск',     TRUE,
        '{"credit_rating":0.25,"ytm":0.20,"ytm_premium":0.15,"duration":0.10,"liquidity":0.10,"category":0.05,"put_offer_soon":0.00,"issue_size":0.05,"coupon_type":0.05,"rating_age":0.05,"dohod_quality":0.00,"dohod_stability":0.00}'::JSONB),
    ('high', 'Повышенный риск',  TRUE,
        '{"credit_rating":0.10,"ytm":0.30,"ytm_premium":0.25,"duration":0.05,"liquidity":0.05,"category":0.05,"put_offer_soon":0.00,"issue_size":0.05,"coupon_type":0.05,"rating_age":0.05,"dohod_quality":0.05,"dohod_stability":0.05}'::JSONB)
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Phase 4 — portfolio (reserved table)
-- ============================================================================
CREATE TABLE IF NOT EXISTS portfolio_positions (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secid      VARCHAR(64)  NOT NULL,
    qty        BIGINT       NOT NULL,
    price_in   DOUBLE PRECISION,
    date_in    DATE,
    note       TEXT         NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_portfolio_positions_user ON portfolio_positions (user_id);

DROP TRIGGER IF EXISTS trg_portfolio_positions_updated ON portfolio_positions;
CREATE TRIGGER trg_portfolio_positions_updated BEFORE UPDATE ON portfolio_positions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
