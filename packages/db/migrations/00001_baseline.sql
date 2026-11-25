-- +goose Up
CREATE TABLE users (
    id uuid CONSTRAINT users_pkey PRIMARY KEY,
    email text NOT NULL,
    email_normalized text NOT NULL CONSTRAINT users_email_normalized_key UNIQUE,
    display_name text NOT NULL,
    status text NOT NULL CONSTRAINT users_status_check CHECK (status IN ('active','disabled')),
    email_verified_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT users_email_check CHECK (email = btrim(email) AND email ~ '^[^[:space:]@]+@[^[:space:]@]+$'),
    CONSTRAINT users_email_normalized_check CHECK (email_normalized = lower(btrim(email)) AND email_normalized ~ '^[^[:space:]@]+@[^[:space:]@]+$'),
    CONSTRAINT users_display_name_check CHECK (btrim(display_name) <> ''),
    CONSTRAINT users_times_check CHECK (isfinite(created_at) AND isfinite(updated_at) AND updated_at >= created_at AND (email_verified_at IS NULL OR isfinite(email_verified_at) AND email_verified_at >= created_at))
);
CREATE TABLE password_credentials (
    user_id uuid CONSTRAINT password_credentials_pkey PRIMARY KEY,
    password_hash text NOT NULL CONSTRAINT password_credentials_hash_check CHECK (btrim(password_hash) <> ''),
    password_changed_at timestamptz NOT NULL CONSTRAINT password_credentials_time_check CHECK (isfinite(password_changed_at)),
    CONSTRAINT password_credentials_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE TABLE email_action_tokens (
    id uuid CONSTRAINT email_action_tokens_pkey PRIMARY KEY,
    user_id uuid NOT NULL,
    purpose text NOT NULL CONSTRAINT email_action_tokens_purpose_check CHECK (purpose IN ('verify_email','reset_password')),
    target_email_normalized text NOT NULL CONSTRAINT email_action_tokens_email_check CHECK (target_email_normalized = lower(btrim(target_email_normalized)) AND target_email_normalized ~ '^[^[:space:]@]+@[^[:space:]@]+$'),
    token_hash bytea NOT NULL CONSTRAINT email_action_tokens_hash_key UNIQUE CONSTRAINT email_action_tokens_hash_check CHECK (octet_length(token_hash) > 0),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    CONSTRAINT email_action_tokens_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT email_action_tokens_times_check CHECK (isfinite(created_at) AND isfinite(expires_at) AND expires_at > created_at AND (consumed_at IS NULL OR isfinite(consumed_at) AND consumed_at >= created_at) AND (revoked_at IS NULL OR isfinite(revoked_at) AND revoked_at >= created_at) AND NOT (consumed_at IS NOT NULL AND revoked_at IS NOT NULL))
);
CREATE INDEX email_action_tokens_user_idx ON email_action_tokens(user_id);
CREATE INDEX email_action_tokens_expiry_idx ON email_action_tokens(expires_at);
CREATE TABLE auth_sessions (
    id uuid CONSTRAINT auth_sessions_pkey PRIMARY KEY,
    user_id uuid NOT NULL,
    secret_hash bytea NOT NULL CONSTRAINT auth_sessions_hash_key UNIQUE CONSTRAINT auth_sessions_hash_check CHECK (octet_length(secret_hash) > 0),
    authenticated_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    CONSTRAINT auth_sessions_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT auth_sessions_id_user_key UNIQUE (id,user_id),
    CONSTRAINT auth_sessions_times_check CHECK (isfinite(created_at) AND isfinite(authenticated_at) AND isfinite(last_seen_at) AND isfinite(expires_at) AND authenticated_at >= created_at AND last_seen_at >= authenticated_at AND expires_at > authenticated_at AND (revoked_at IS NULL OR isfinite(revoked_at) AND revoked_at >= created_at))
);
CREATE INDEX auth_sessions_user_idx ON auth_sessions(user_id);
CREATE INDEX auth_sessions_expiry_idx ON auth_sessions(expires_at);
CREATE TABLE organizations (
    id uuid CONSTRAINT organizations_pkey PRIMARY KEY,
    name text NOT NULL CONSTRAINT organizations_name_check CHECK (btrim(name) <> ''),
    slug text NOT NULL CONSTRAINT organizations_slug_key UNIQUE CONSTRAINT organizations_slug_check CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    status text NOT NULL CONSTRAINT organizations_status_check CHECK (status IN ('active','suspended')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    logo_object_key text CONSTRAINT organizations_logo_check CHECK (logo_object_key IS NULL OR logo_object_key ~ '^[A-Za-z0-9/_.-]{1,512}$'),
    CONSTRAINT organizations_times_check CHECK (isfinite(created_at) AND isfinite(updated_at) AND updated_at >= created_at)
);
CREATE TABLE organization_memberships (
    id uuid CONSTRAINT organization_memberships_pkey PRIMARY KEY,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status text NOT NULL CONSTRAINT organization_memberships_status_check CHECK (status IN ('active','removed')),
    role text NOT NULL CONSTRAINT organization_memberships_role_check CHECK (role IN ('owner','admin','member')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT organization_memberships_organization_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_org_user_key UNIQUE (organization_id,user_id),
    CONSTRAINT organization_memberships_id_org_key UNIQUE (id,organization_id),
    CONSTRAINT organization_memberships_owner_check CHECK (role <> 'owner' OR status = 'active'),
    CONSTRAINT organization_memberships_times_check CHECK (isfinite(created_at) AND isfinite(updated_at) AND updated_at >= created_at)
);
CREATE INDEX organization_memberships_user_idx ON organization_memberships(user_id,organization_id);
CREATE INDEX organization_memberships_owners_idx ON organization_memberships(organization_id,user_id) WHERE role = 'owner' AND status='active';
CREATE TABLE platform_staff (
    id uuid CONSTRAINT platform_staff_pkey PRIMARY KEY,
    status text NOT NULL CONSTRAINT platform_staff_status_check CHECK (status IN ('active','disabled')),
    role text NOT NULL CONSTRAINT platform_staff_role_check CHECK (role IN ('superadmin','admin')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    email_normalized text NOT NULL CONSTRAINT platform_staff_email_check CHECK (email_normalized=lower(btrim(email_normalized)) AND email_normalized ~ '^[^[:space:]@]+@[^[:space:]@]+$') CONSTRAINT platform_staff_email_key UNIQUE,
    display_name text NOT NULL CONSTRAINT platform_staff_name_check CHECK (btrim(display_name)<>''),
    CONSTRAINT platform_staff_times_check CHECK (isfinite(created_at) AND isfinite(updated_at) AND updated_at >= created_at)
);

-- +goose StatementBegin
CREATE FUNCTION audit_detail_valid(detail jsonb) RETURNS boolean
LANGUAGE plpgsql IMMUTABLE STRICT AS $$
DECLARE item record; field jsonb;
BEGIN
 IF jsonb_typeof(detail)<>'object' OR octet_length(detail::text)>16384 THEN RETURN false; END IF;
 FOR item IN SELECT * FROM jsonb_each(detail) LOOP
  IF item.key='reason' THEN
   IF jsonb_typeof(item.value)<>'string' OR length(item.value #>> '{}')>1024 THEN RETURN false; END IF;
  ELSIF item.key='changed_fields' THEN
   IF jsonb_typeof(item.value)<>'array' THEN RETURN false; END IF;
   IF jsonb_array_length(item.value)>64 THEN RETURN false; END IF;
   FOR field IN SELECT * FROM jsonb_array_elements(item.value) LOOP
    IF jsonb_typeof(field)<>'string' OR (field #>> '{}') !~ '^[a-z][a-z0-9_]{0,63}$' THEN RETURN false; END IF;
   END LOOP;
  ELSIF item.key IN ('previous','current') THEN
   IF jsonb_typeof(item.value)<>'object' THEN RETURN false; END IF;
   IF EXISTS(SELECT 1 FROM jsonb_each(item.value) AS v WHERE
    v.key NOT IN ('status','role') OR
    jsonb_typeof(v.value) NOT IN ('string','boolean','number','null') OR
    (v.key='role' AND v.value NOT IN ('"superadmin"'::jsonb,'"admin"'::jsonb))) THEN RETURN false; END IF;
  ELSE RETURN false;
  END IF;
 END LOOP;
 RETURN true;
END $$;
-- +goose StatementEnd
-- Structural storage allowlist only, not an unrestricted event payload API.
-- Domain event allowlists and secret exclusion remain caller responsibilities.
CREATE TABLE audit_events (
    id uuid CONSTRAINT audit_events_pkey PRIMARY KEY,
    occurred_at timestamptz NOT NULL CONSTRAINT audit_events_time_check CHECK (isfinite(occurred_at)),
    actor_kind text NOT NULL,
    actor_user_id uuid,
    actor_staff_id uuid,
    organization_id uuid,
    action text NOT NULL CONSTRAINT audit_events_action_check CHECK (length(action)<=128 AND action ~ '^[a-z][a-z0-9_.]*$'),
    outcome text NOT NULL CONSTRAINT audit_events_outcome_check CHECK (outcome IN ('success','failure','denied')),
    target_type text,
    target_id text,
    correlation_id uuid,
    detail jsonb NOT NULL CONSTRAINT audit_events_detail_check CHECK (audit_detail_valid(detail)),
    CONSTRAINT audit_events_actor_check CHECK (
        (actor_kind='user' AND actor_user_id IS NOT NULL AND actor_staff_id IS NULL)
        OR (actor_kind='staff' AND actor_user_id IS NULL AND actor_staff_id IS NOT NULL)
        OR (actor_kind IN ('system','anonymous') AND actor_user_id IS NULL AND actor_staff_id IS NULL)
    ),
    CONSTRAINT audit_events_target_check CHECK ((target_type IS NULL AND target_id IS NULL) OR (target_type IS NOT NULL AND target_id IS NOT NULL AND length(target_type)<=64 AND target_type ~ '^[a-z][a-z0-9_]*$' AND length(target_id) BETWEEN 1 AND 256)),
    CONSTRAINT audit_events_actor_user_fkey FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT audit_events_actor_staff_fkey FOREIGN KEY (actor_staff_id) REFERENCES platform_staff(id) ON DELETE RESTRICT,
    CONSTRAINT audit_events_organization_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT
);
CREATE INDEX audit_events_org_time_idx ON audit_events(organization_id,occurred_at);
CREATE INDEX audit_events_actor_user_idx ON audit_events(actor_user_id);
CREATE INDEX audit_events_actor_staff_idx ON audit_events(actor_staff_id);
CREATE INDEX audit_events_time_idx ON audit_events(occurred_at);
-- +goose StatementBegin
CREATE FUNCTION audit_events_append_only() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'audit events are append only' USING ERRCODE='42501';
END $$;
-- +goose StatementEnd
CREATE TRIGGER audit_events_append_only BEFORE UPDATE OR DELETE OR TRUNCATE ON audit_events FOR EACH STATEMENT EXECUTE FUNCTION audit_events_append_only();


-- Write fences, not lock-only reads: stale Repeatable Read writers must abort.
-- +goose StatementBegin
CREATE FUNCTION core_owner_fence() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE uid uuid; oid uuid;
BEGIN
  IF TG_TABLE_NAME = 'organization_memberships' THEN
    IF TG_OP = 'UPDATE' AND (NEW.id,NEW.user_id,NEW.organization_id) IS DISTINCT FROM (OLD.id,OLD.user_id,OLD.organization_id) THEN
      RAISE EXCEPTION 'membership identity is immutable' USING ERRCODE='23514', CONSTRAINT='organization_owner_identity_check';
    END IF;
    IF TG_OP = 'DELETE' THEN uid := OLD.user_id; oid := OLD.organization_id;
    ELSE uid := NEW.user_id; oid := NEW.organization_id; END IF;
    UPDATE users SET updated_at=updated_at WHERE id=uid;
    UPDATE organizations SET updated_at=updated_at WHERE id=oid;
  ELSIF TG_TABLE_NAME = 'users' THEN
    IF TG_OP = 'UPDATE' AND NEW.id IS DISTINCT FROM OLD.id THEN
      RAISE EXCEPTION 'user identity is immutable' USING ERRCODE='23514', CONSTRAINT='organization_owner_identity_check';
    END IF;
    FOR oid IN SELECT organization_id FROM organization_memberships WHERE user_id=OLD.id ORDER BY organization_id LOOP
      UPDATE organizations SET updated_at=updated_at WHERE id=oid;
    END LOOP;
  ELSE
    IF NEW.id IS DISTINCT FROM OLD.id THEN
      RAISE EXCEPTION 'organization identity is immutable' USING ERRCODE='23514', CONSTRAINT='organization_owner_identity_check';
    END IF;
  END IF;
  IF TG_OP='DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END $$;
CREATE FUNCTION core_check_owner(oid uuid) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  IF EXISTS (SELECT 1 FROM organizations WHERE id=oid) AND NOT EXISTS (
    SELECT 1 FROM organization_memberships m JOIN users u ON u.id=m.user_id
    WHERE m.organization_id=oid AND m.status='active' AND m.role = 'owner' AND u.status='active'
  ) THEN
    RAISE EXCEPTION 'organization requires an active owner' USING ERRCODE='23514', CONSTRAINT='organization_active_owner_check';
  END IF;
END $$;
CREATE FUNCTION core_owner_validate() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE oid uuid;
BEGIN
  IF TG_TABLE_NAME='organizations' THEN PERFORM core_check_owner(NEW.id);
  ELSIF TG_TABLE_NAME='organization_memberships' THEN
    IF TG_OP='DELETE' THEN PERFORM core_check_owner(OLD.organization_id);
    ELSE PERFORM core_check_owner(NEW.organization_id); END IF;
  ELSE
    FOR oid IN SELECT organization_id FROM organization_memberships WHERE user_id=OLD.id ORDER BY organization_id LOOP
      PERFORM core_check_owner(oid);
    END LOOP;
  END IF;
  RETURN NULL;
END $$;
CREATE FUNCTION core_owner_no_truncate() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'owner relations cannot be truncated' USING ERRCODE='23514', CONSTRAINT='organization_active_owner_check';
END $$;
CREATE TRIGGER organization_memberships_owner_fence BEFORE INSERT OR UPDATE OR DELETE ON organization_memberships FOR EACH ROW EXECUTE FUNCTION core_owner_fence();
CREATE TRIGGER users_owner_fence BEFORE UPDATE OF id,status OR DELETE ON users FOR EACH ROW EXECUTE FUNCTION core_owner_fence();
CREATE TRIGGER organizations_owner_identity BEFORE UPDATE OF id ON organizations FOR EACH ROW EXECUTE FUNCTION core_owner_fence();
CREATE CONSTRAINT TRIGGER organizations_owner_validate AFTER INSERT ON organizations DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION core_owner_validate();
CREATE CONSTRAINT TRIGGER organization_memberships_owner_validate AFTER INSERT OR UPDATE OR DELETE ON organization_memberships DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION core_owner_validate();
CREATE CONSTRAINT TRIGGER users_owner_validate AFTER UPDATE OF status OR DELETE ON users DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION core_owner_validate();
CREATE TRIGGER organization_memberships_no_truncate BEFORE TRUNCATE ON organization_memberships FOR EACH STATEMENT EXECUTE FUNCTION core_owner_no_truncate();
CREATE TRIGGER users_owner_no_truncate BEFORE TRUNCATE ON users FOR EACH STATEMENT EXECUTE FUNCTION core_owner_no_truncate();
DO $$ DECLARE oid uuid; BEGIN
  FOR oid IN SELECT id FROM organizations ORDER BY id LOOP PERFORM core_check_owner(oid); END LOOP;
END $$;
-- +goose StatementEnd


CREATE TABLE platform_staff_credentials (
 staff_id uuid CONSTRAINT platform_staff_credentials_pkey PRIMARY KEY,
 password_hash text NOT NULL CONSTRAINT platform_staff_credentials_hash_check CHECK (length(password_hash) BETWEEN 80 AND 256 AND password_hash LIKE '$argon2id$v=19$%'),
 changed_at timestamptz NOT NULL CONSTRAINT platform_staff_credentials_time_check CHECK (isfinite(changed_at)),
 CONSTRAINT platform_staff_credentials_staff_fkey FOREIGN KEY(staff_id) REFERENCES platform_staff(id) ON DELETE RESTRICT
);
CREATE TABLE platform_staff_sessions (
 id uuid CONSTRAINT platform_staff_sessions_pkey PRIMARY KEY,
 staff_id uuid NOT NULL,
 token_hash bytea NOT NULL CONSTRAINT platform_staff_sessions_hash_key UNIQUE CONSTRAINT platform_staff_sessions_hash_check CHECK (octet_length(token_hash)=32),
 created_at timestamptz NOT NULL,
 expires_at timestamptz NOT NULL,
 revoked_at timestamptz,
 CONSTRAINT platform_staff_sessions_staff_fkey FOREIGN KEY(staff_id) REFERENCES platform_staff(id) ON DELETE RESTRICT,
 CONSTRAINT platform_staff_sessions_times_check CHECK (isfinite(created_at) AND isfinite(expires_at) AND expires_at-created_at=interval '720 hours' AND (revoked_at IS NULL OR isfinite(revoked_at) AND revoked_at>=created_at))
);
CREATE INDEX platform_staff_sessions_staff_idx ON platform_staff_sessions(staff_id);
CREATE INDEX platform_staff_sessions_expiry_idx ON platform_staff_sessions(expires_at);
CREATE TABLE platform_staff_auth_control (
 id smallint CONSTRAINT platform_staff_auth_control_pkey PRIMARY KEY CONSTRAINT platform_staff_auth_control_singleton_check CHECK (id=1),
 mutation_epoch bigint NOT NULL CONSTRAINT platform_staff_auth_control_epoch_check CHECK (mutation_epoch>=0),
 bootstrapped_at timestamptz CONSTRAINT platform_staff_auth_control_time_check CHECK (bootstrapped_at IS NULL OR isfinite(bootstrapped_at))
);
INSERT INTO platform_staff_auth_control VALUES (1,0,NULL);
CREATE TABLE platform_staff_auth_buckets (
 purpose text NOT NULL CONSTRAINT platform_staff_auth_buckets_purpose_check CHECK (purpose IN ('login','confirmation')),
 dimension text NOT NULL CONSTRAINT platform_staff_auth_buckets_dimension_check CHECK (dimension IN ('account','source')),
 key_hash bytea NOT NULL CONSTRAINT platform_staff_auth_buckets_hash_check CHECK (octet_length(key_hash)=32),
 window_start timestamptz NOT NULL,
 attempts bigint NOT NULL CONSTRAINT platform_staff_auth_buckets_attempts_check CHECK (attempts>=0),
 expires_at timestamptz NOT NULL,
 failure_recorded boolean NOT NULL DEFAULT false,
 CONSTRAINT platform_staff_auth_buckets_pkey PRIMARY KEY(purpose,dimension,key_hash,window_start),
 CONSTRAINT platform_staff_auth_buckets_times_check CHECK (isfinite(window_start) AND isfinite(expires_at) AND expires_at=window_start+interval '15 minutes')
);
CREATE INDEX platform_staff_auth_buckets_expiry_idx ON platform_staff_auth_buckets(expires_at);

-- +goose StatementBegin
CREATE FUNCTION core_staff_read_only() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'staff history is retained read-only' USING ERRCODE='42501';
END $$;
CREATE FUNCTION core_staff_fence() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 -- Statement-level fence runs BEFORE target row locks. Actual writes reject
 -- stale Repeatable Read snapshots (40001), not just concurrent RC writers.
 UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1;
 IF NOT FOUND THEN RAISE EXCEPTION 'staff control missing' USING ERRCODE='23514'; END IF;
 RETURN NULL;
END $$;
CREATE FUNCTION core_staff_control_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.id<>OLD.id OR NEW.mutation_epoch<OLD.mutation_epoch OR
  (OLD.bootstrapped_at IS NOT NULL AND NEW.bootstrapped_at IS DISTINCT FROM OLD.bootstrapped_at) THEN
  RAISE EXCEPTION 'staff bootstrap latch is irreversible' USING ERRCODE='23514',CONSTRAINT='platform_staff_bootstrap_latch_check';
 END IF;
 RETURN NEW;
END $$;
CREATE FUNCTION core_staff_changed() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='UPDATE' AND NEW.id IS DISTINCT FROM OLD.id THEN
  RAISE EXCEPTION 'staff identity is immutable' USING ERRCODE='23514',CONSTRAINT='platform_staff_identity_check';
 END IF;
 IF NEW.role='superadmin' AND NEW.status='active' THEN
  UPDATE platform_staff_auth_control SET bootstrapped_at=coalesce(bootstrapped_at,statement_timestamp()) WHERE id=1;
 END IF;
 IF TG_OP='UPDATE' AND (NEW.role,NEW.status) IS DISTINCT FROM (OLD.role,OLD.status) THEN
  UPDATE platform_staff_sessions SET revoked_at=greatest(created_at,statement_timestamp()) WHERE staff_id=NEW.id AND revoked_at IS NULL;
 END IF;
 RETURN NEW;
END $$;
CREATE FUNCTION core_staff_credential_changed() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='UPDATE' AND NEW.staff_id IS DISTINCT FROM OLD.staff_id THEN
  RAISE EXCEPTION 'credential owner is immutable' USING ERRCODE='23514';
 END IF;
 IF TG_OP='DELETE' THEN
  UPDATE platform_staff_sessions SET revoked_at=greatest(created_at,statement_timestamp()) WHERE staff_id=OLD.staff_id AND revoked_at IS NULL;
  RETURN OLD;
 END IF;
 UPDATE platform_staff_sessions SET revoked_at=greatest(created_at,statement_timestamp()) WHERE staff_id=NEW.staff_id AND revoked_at IS NULL;
 RETURN NEW;
END $$;
CREATE FUNCTION core_staff_validate() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF EXISTS(SELECT 1 FROM platform_staff_auth_control WHERE id=1 AND bootstrapped_at IS NOT NULL) AND
  NOT EXISTS(SELECT 1 FROM platform_staff WHERE role='superadmin' AND status='active') THEN
  RAISE EXCEPTION 'one active superadmin must remain' USING ERRCODE='23514',CONSTRAINT='platform_staff_last_superadmin_check';
 END IF;
 RETURN NULL;
END $$;
CREATE FUNCTION core_staff_session_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='UPDATE' AND ((NEW.id,NEW.staff_id,NEW.token_hash,NEW.created_at,NEW.expires_at) IS DISTINCT FROM (OLD.id,OLD.staff_id,OLD.token_hash,OLD.created_at,OLD.expires_at) OR
  OLD.revoked_at IS NOT NULL AND NEW.revoked_at IS DISTINCT FROM OLD.revoked_at) THEN
  RAISE EXCEPTION 'session identity and revocation are immutable' USING ERRCODE='23514',CONSTRAINT='platform_staff_sessions_immutable_check';
 END IF;
 IF TG_OP='INSERT' AND NOT EXISTS(SELECT 1 FROM platform_staff s JOIN platform_staff_credentials c ON c.staff_id=s.id WHERE s.id=NEW.staff_id AND s.status='active') THEN
  RAISE EXCEPTION 'session requires active independent credentials' USING ERRCODE='23514',CONSTRAINT='platform_staff_sessions_active_check';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER platform_staff_fence BEFORE INSERT OR UPDATE OR DELETE ON platform_staff FOR EACH STATEMENT EXECUTE FUNCTION core_staff_fence();
CREATE TRIGGER platform_staff_no_delete BEFORE DELETE OR TRUNCATE ON platform_staff FOR EACH STATEMENT EXECUTE FUNCTION core_staff_read_only();
CREATE TRIGGER platform_staff_changed BEFORE INSERT OR UPDATE ON platform_staff FOR EACH ROW EXECUTE FUNCTION core_staff_changed();
CREATE CONSTRAINT TRIGGER platform_staff_validate AFTER INSERT OR UPDATE ON platform_staff DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION core_staff_validate();
CREATE TRIGGER platform_staff_control_guard BEFORE UPDATE ON platform_staff_auth_control FOR EACH ROW EXECUTE FUNCTION core_staff_control_guard();
CREATE TRIGGER platform_staff_control_retained BEFORE DELETE OR TRUNCATE ON platform_staff_auth_control FOR EACH STATEMENT EXECUTE FUNCTION core_staff_read_only();
CREATE CONSTRAINT TRIGGER platform_staff_control_validate AFTER INSERT OR UPDATE ON platform_staff_auth_control DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION core_staff_validate();
CREATE TRIGGER platform_staff_credentials_fence BEFORE INSERT OR UPDATE OR DELETE ON platform_staff_credentials FOR EACH STATEMENT EXECUTE FUNCTION core_staff_fence();
CREATE TRIGGER platform_staff_credentials_changed BEFORE INSERT OR UPDATE OR DELETE ON platform_staff_credentials FOR EACH ROW EXECUTE FUNCTION core_staff_credential_changed();
CREATE TRIGGER platform_staff_credentials_no_truncate BEFORE TRUNCATE ON platform_staff_credentials FOR EACH STATEMENT EXECUTE FUNCTION core_staff_read_only();
CREATE TRIGGER platform_staff_sessions_fence BEFORE INSERT OR UPDATE OR DELETE ON platform_staff_sessions FOR EACH STATEMENT EXECUTE FUNCTION core_staff_fence();
CREATE TRIGGER platform_staff_sessions_guard BEFORE INSERT OR UPDATE ON platform_staff_sessions FOR EACH ROW EXECUTE FUNCTION core_staff_session_guard();


-- +goose StatementEnd


-- Nullable generations deliberately leave historical users/material ineligible
-- for the new issuer until an application command establishes current proof.
ALTER TABLE users
    ADD COLUMN telephone text,
    ADD COLUMN identity_version bigint,
    ADD CONSTRAINT users_telephone_check CHECK (telephone IS NULL OR telephone ~ '^\+[1-9][0-9]{1,14}$'),
    ADD CONSTRAINT users_identity_version_check CHECK (identity_version > 0);

CREATE TABLE identity_enrollment_receipts (
    secret_hash bytea CONSTRAINT identity_enrollment_receipts_pkey PRIMARY KEY,
    user_id uuid,
    identity_version bigint,
    credential_generation bigint,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    CONSTRAINT identity_enrollment_receipts_hash_check CHECK (octet_length(secret_hash)=32),
    CONSTRAINT identity_enrollment_receipts_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT identity_enrollment_receipts_binding_check CHECK (
        (user_id IS NULL AND identity_version IS NULL AND credential_generation IS NULL)
        OR (user_id IS NOT NULL AND identity_version IS NOT NULL AND identity_version>0
            AND credential_generation IS NOT NULL AND credential_generation>0)),
    CONSTRAINT identity_enrollment_receipts_times_check CHECK (
        isfinite(created_at) AND isfinite(expires_at) AND expires_at>created_at
        AND expires_at<=created_at+interval '30 minutes')
);
CREATE INDEX identity_enrollment_receipts_expiry_idx ON identity_enrollment_receipts(expires_at);
CREATE INDEX identity_enrollment_receipts_user_idx ON identity_enrollment_receipts(user_id);

ALTER TABLE password_credentials
    ADD COLUMN generation bigint,
    ADD CONSTRAINT password_credentials_generation_check CHECK (generation>0);
ALTER TABLE email_action_tokens DROP CONSTRAINT email_action_tokens_purpose_check;
ALTER TABLE email_action_tokens
    ADD COLUMN identity_version bigint,
    ADD COLUMN credential_generation bigint,
    ADD COLUMN enrollment_hash bytea,
    ADD CONSTRAINT email_action_tokens_purpose_check CHECK (purpose IN ('verify_email','setup_password','reset_password')),
    ADD CONSTRAINT email_action_tokens_binding_check CHECK (
        (identity_version IS NULL AND credential_generation IS NULL AND enrollment_hash IS NULL AND purpose<>'setup_password')
        OR (identity_version IS NOT NULL AND identity_version>0 AND credential_generation IS NOT NULL
            AND octet_length(token_hash)=32
            AND ((purpose='setup_password' AND credential_generation=0 AND enrollment_hash IS NULL)
                OR (purpose='reset_password' AND credential_generation>0 AND enrollment_hash IS NULL)
                OR (purpose='verify_email' AND credential_generation>0 AND enrollment_hash IS NOT NULL AND octet_length(enrollment_hash)=32)))) ,
    ADD CONSTRAINT email_action_tokens_lifetime_check CHECK (identity_version IS NULL OR
        expires_at<=created_at+CASE WHEN purpose='reset_password' THEN interval '30 minutes' ELSE interval '24 hours' END);
-- Enrollment hashes are historical proof bindings: expiry/cleanup must not
-- invalidate referential integrity or require deleting append-only token history.
CREATE INDEX email_action_tokens_live_user_purpose_idx ON email_action_tokens(user_id,purpose,expires_at)
    WHERE consumed_at IS NULL AND revoked_at IS NULL;
-- Issuers MUST already hold the user before entering the email-action stage.
-- The real write fence also rejects obsolete Repeatable Read snapshots instead
-- of counting actions from an earlier snapshot after waiting for another issuer.
-- created_at is the application-selected accepted time, sampled after locks.
-- +goose StatementBegin
CREATE FUNCTION identity_email_action_capacity_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET identity_version=identity_version WHERE id=NEW.user_id;
    IF (SELECT count(*) FROM email_action_tokens
        WHERE user_id=NEW.user_id AND purpose=NEW.purpose
        AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>NEW.created_at)>=10 THEN
        RAISE EXCEPTION 'email action capacity unavailable' USING ERRCODE='23514', CONSTRAINT='email_action_tokens_capacity_check';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd
CREATE TRIGGER identity_email_action_capacity_guard BEFORE INSERT ON email_action_tokens
    FOR EACH ROW WHEN (NEW.identity_version IS NOT NULL) EXECUTE FUNCTION identity_email_action_capacity_guard();

ALTER TABLE auth_sessions
    ADD COLUMN kind text,
    ADD COLUMN identity_version bigint,
    ADD COLUMN credential_generation bigint,
    ADD COLUMN last_activity_at timestamptz,
    ADD COLUMN browser_cookie_expires_at timestamptz,
    ADD CONSTRAINT auth_sessions_identity_check CHECK (
        (kind IS NULL AND identity_version IS NULL AND credential_generation IS NULL
            AND last_activity_at IS NULL AND browser_cookie_expires_at IS NULL)
        OR (kind IS NOT NULL AND kind IN ('restricted','full') AND identity_version IS NOT NULL AND identity_version>0
            AND credential_generation IS NOT NULL AND credential_generation>0 AND octet_length(secret_hash)=32
            AND last_activity_at IS NOT NULL AND isfinite(last_activity_at) AND last_activity_at>=authenticated_at
            AND browser_cookie_expires_at IS NOT NULL AND isfinite(browser_cookie_expires_at)
            AND browser_cookie_expires_at>authenticated_at AND browser_cookie_expires_at<=expires_at
            AND ((kind='restricted' AND last_activity_at=authenticated_at AND expires_at<=authenticated_at+interval '30 minutes')
                OR (kind='full' AND expires_at=last_activity_at+interval '1176 hours'))));

-- A tiny persistent write fence makes bucket capacity atomic across replicas.
-- Admission transactions prelock this row, then sorted bucket keys, and finish
-- before entering any user/credential/session mutation lock chain.
CREATE TABLE identity_limiter_control (
    id integer CONSTRAINT identity_limiter_control_pkey PRIMARY KEY,
    bucket_count integer NOT NULL,
    CONSTRAINT identity_limiter_control_singleton_check CHECK (id=1),
    CONSTRAINT identity_limiter_control_count_check CHECK (bucket_count BETWEEN 0 AND 100000)
);
INSERT INTO identity_limiter_control(id,bucket_count) VALUES(1,0);
CREATE TABLE identity_limiter_buckets (
    key_hash bytea CONSTRAINT identity_limiter_buckets_pkey PRIMARY KEY,
    window_started_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    attempts integer NOT NULL,
    CONSTRAINT identity_limiter_buckets_hash_check CHECK (octet_length(key_hash)=32),
    CONSTRAINT identity_limiter_buckets_attempts_check CHECK (attempts>0),
    CONSTRAINT identity_limiter_buckets_times_check CHECK (
        isfinite(window_started_at) AND isfinite(expires_at) AND expires_at>window_started_at
        AND expires_at<=window_started_at+interval '24 hours')
);
CREATE INDEX identity_limiter_buckets_expiry_idx ON identity_limiter_buckets(expires_at);
-- +goose StatementBegin
CREATE FUNCTION identity_limiter_capacity_guard() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP='TRUNCATE' THEN
        RAISE EXCEPTION 'limiter truncation forbidden' USING ERRCODE='23514', CONSTRAINT='identity_limiter_capacity_check';
    ELSIF TG_OP='INSERT' THEN
        UPDATE identity_limiter_control SET bucket_count=bucket_count+1 WHERE id=1 AND bucket_count<100000;
    ELSE
        UPDATE identity_limiter_control SET bucket_count=bucket_count-1 WHERE id=1 AND bucket_count>0;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'limiter capacity unavailable' USING ERRCODE='23514', CONSTRAINT='identity_limiter_capacity_check';
    END IF;
    RETURN NULL;
END $$;
-- +goose StatementEnd
CREATE TRIGGER identity_limiter_capacity_guard AFTER INSERT OR DELETE ON identity_limiter_buckets
    FOR EACH ROW EXECUTE FUNCTION identity_limiter_capacity_guard();
CREATE TRIGGER identity_limiter_no_truncate BEFORE TRUNCATE ON identity_limiter_buckets
    FOR EACH STATEMENT EXECUTE FUNCTION identity_limiter_capacity_guard();


-- Organization API keys for the public API. Only a SHA-256 token hash is stored;
-- the prefix identifies a key in UIs and logs without revealing it.
CREATE TABLE api_keys (
    id uuid CONSTRAINT api_keys_pkey PRIMARY KEY,
    organization_id uuid NOT NULL,
    name text NOT NULL CONSTRAINT api_keys_name_check CHECK (btrim(name) <> '' AND length(name) <= 128),
    token_prefix text NOT NULL CONSTRAINT api_keys_prefix_check CHECK (token_prefix ~ '^zk_[A-Za-z0-9]{4,12}$'),
    token_hash bytea NOT NULL CONSTRAINT api_keys_hash_key UNIQUE CONSTRAINT api_keys_hash_check CHECK (octet_length(token_hash) = 32),
    last_used_at timestamptz,
    expires_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL,
    CONSTRAINT api_keys_organization_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE RESTRICT,
    CONSTRAINT api_keys_times_check CHECK (isfinite(created_at) AND (last_used_at IS NULL OR isfinite(last_used_at) AND last_used_at >= created_at) AND (expires_at IS NULL OR isfinite(expires_at) AND expires_at > created_at) AND (revoked_at IS NULL OR isfinite(revoked_at) AND revoked_at >= created_at))
);
CREATE INDEX api_keys_organization_idx ON api_keys(organization_id);


-- River v0.51.2 canonical migrations from river migrate-get --all --up.
-- River main migration 001 [up]
-- The fresh baseline stores River's final migration catalog shape, visible to
-- offline sqlc; canonical migration 005 still rebuilds it when applied.
CREATE TABLE river_migration(
  line text NOT NULL,
  version bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  CONSTRAINT line_length CHECK (char_length(line) > 0 AND char_length(line) < 128),
  CONSTRAINT version_gte_1 CHECK (version >= 1),
  PRIMARY KEY (line, version)
);

-- River main migration 002 [up]
-- Goose applies this entire baseline in one transaction, so define River's later pending value up front.
CREATE TYPE river_job_state AS ENUM(
  'available',
  'cancelled',
  'completed',
  'discarded',
  'pending',
  'retryable',
  'running',
  'scheduled'
);

CREATE TABLE river_job(
  -- 8 bytes
  id bigserial PRIMARY KEY,

  -- 8 bytes (4 bytes + 2 bytes + 2 bytes)
  --
  -- `state` is kept near the top of the table for operator convenience -- when
  -- looking at jobs with `SELECT *` it'll appear first after ID. The other two
  -- fields aren't as important but are kept adjacent to `state` for alignment
  -- to get an 8-byte block.
  state river_job_state NOT NULL DEFAULT 'available',
  attempt smallint NOT NULL DEFAULT 0,
  max_attempts smallint NOT NULL,

  -- 8 bytes each (no alignment needed)
  attempted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  finalized_at timestamptz,
  scheduled_at timestamptz NOT NULL DEFAULT NOW(),

  -- 2 bytes (some wasted padding probably)
  priority smallint NOT NULL DEFAULT 1,

  -- types stored out-of-band
  args jsonb,
  attempted_by text[],
  errors jsonb[],
  kind text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}',
  queue text NOT NULL DEFAULT 'default',
  tags varchar(255)[],

  CONSTRAINT finalized_or_finalized_at_null CHECK ((state IN ('cancelled', 'completed', 'discarded') AND finalized_at IS NOT NULL) OR finalized_at IS NULL),
  CONSTRAINT max_attempts_is_positive CHECK (max_attempts > 0),
  CONSTRAINT priority_in_range CHECK (priority >= 1 AND priority <= 4),
  CONSTRAINT queue_length CHECK (char_length(queue) > 0 AND char_length(queue) < 128),
  CONSTRAINT kind_length CHECK (char_length(kind) > 0 AND char_length(kind) < 128)
);

-- We may want to consider adding another property here after `kind` if it seems
-- like it'd be useful for something.
CREATE INDEX river_job_kind ON river_job USING btree(kind);

CREATE INDEX river_job_state_and_finalized_at_index ON river_job USING btree(state, finalized_at) WHERE finalized_at IS NOT NULL;

CREATE INDEX river_job_prioritized_fetching_index ON river_job USING btree(state, queue, priority, scheduled_at, id);

CREATE INDEX river_job_args_index ON river_job USING GIN(args);

CREATE INDEX river_job_metadata_index ON river_job USING GIN(metadata);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION river_job_notify()
  RETURNS TRIGGER
  AS $$
DECLARE
  payload json;
BEGIN
  IF NEW.state = 'available' THEN
    -- Notify will coalesce duplicate notifications within a transaction, so
    -- keep these payloads generalized:
    payload = json_build_object('queue', NEW.queue);
    PERFORM
      pg_notify('river_insert', payload::text);
  END IF;
  RETURN NULL;
END;
$$
LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER river_notify
  AFTER INSERT ON river_job
  FOR EACH ROW
  EXECUTE PROCEDURE river_job_notify();

CREATE UNLOGGED TABLE river_leader(
    -- 8 bytes each (no alignment needed)
    elected_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,

    -- types stored out-of-band
    leader_id text NOT NULL,
    name text PRIMARY KEY,

    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT leader_id_length CHECK (char_length(leader_id) > 0 AND char_length(leader_id) < 128)
);

-- River main migration 003 [up]
ALTER TABLE river_job ALTER COLUMN tags SET DEFAULT '{}';
UPDATE river_job SET tags = '{}' WHERE tags IS NULL;
ALTER TABLE river_job ALTER COLUMN tags SET NOT NULL;

-- River main migration 004 [up]
-- The args column never had a NOT NULL constraint or default value at the
-- database level, though we tried to ensure one at the application level.
ALTER TABLE river_job ALTER COLUMN args SET DEFAULT '{}';
UPDATE river_job SET args = '{}' WHERE args IS NULL;
ALTER TABLE river_job ALTER COLUMN args SET NOT NULL;
ALTER TABLE river_job ALTER COLUMN args DROP DEFAULT;

-- The metadata column never had a NOT NULL constraint or default value at the
-- database level, though we tried to ensure one at the application level.
ALTER TABLE river_job ALTER COLUMN metadata SET DEFAULT '{}';
UPDATE river_job SET metadata = '{}' WHERE metadata IS NULL;
ALTER TABLE river_job ALTER COLUMN metadata SET NOT NULL;

-- The 'pending' job state will be used for upcoming functionality:
ALTER TABLE river_job DROP CONSTRAINT finalized_or_finalized_at_null;
ALTER TABLE river_job ADD CONSTRAINT finalized_or_finalized_at_null CHECK (
    (finalized_at IS NULL AND state NOT IN ('cancelled', 'completed', 'discarded')) OR
    (finalized_at IS NOT NULL AND state IN ('cancelled', 'completed', 'discarded'))
);

DROP TRIGGER river_notify ON river_job;
DROP FUNCTION river_job_notify;

--
-- Create table `river_queue`.
--

CREATE TABLE river_queue (
    name text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}' ::jsonb,
    paused_at timestamptz,
    updated_at timestamptz NOT NULL
);

--
-- Alter `river_leader` to add a default value of 'default` to `name`.
--

ALTER TABLE river_leader
    ALTER COLUMN name SET DEFAULT 'default',
    DROP CONSTRAINT name_length,
    ADD CONSTRAINT name_length CHECK (name = 'default');

-- River main migration 005 [up]
--
-- Rebuild the migration table so it's based on `(line, version)`.
--

-- +goose StatementBegin
DO
$body$
BEGIN
    -- Tolerate users who may be using their own migration system rather than
    -- River's. If they are, they will have skipped version 001 containing
    -- `CREATE TABLE river_migration`, so this table won't exist.
    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN
        ALTER TABLE river_migration
            RENAME TO river_migration_old;

        CREATE TABLE river_migration(
            line TEXT NOT NULL,
            version bigint NOT NULL,
            created_at timestamptz NOT NULL DEFAULT NOW(),
            CONSTRAINT line_length CHECK (char_length(line) > 0 AND char_length(line) < 128),
            CONSTRAINT version_gte_1 CHECK (version >= 1),
            PRIMARY KEY (line, version)
        );

        INSERT INTO river_migration
            (created_at, line, version)
        SELECT created_at, 'main', version
        FROM river_migration_old;

        DROP TABLE river_migration_old;
    END IF;
END;
$body$
LANGUAGE 'plpgsql';
-- +goose StatementEnd

--
-- Add `river_job.unique_key` and bring up an index on it.
--

-- These statements use `IF NOT EXISTS` to allow users with a `river_job` table
-- of non-trivial size to build the index `CONCURRENTLY` out of band of this
-- migration, then follow by completing the migration.
ALTER TABLE river_job
    ADD COLUMN IF NOT EXISTS unique_key bytea;

CREATE UNIQUE INDEX IF NOT EXISTS river_job_kind_unique_key_idx ON river_job (kind, unique_key) WHERE unique_key IS NOT NULL;

--
-- Create `river_client` and derivative.
--
-- This feature hasn't quite yet been implemented, but we're taking advantage of
-- the migration to add the schema early so that we can add it later without an
-- additional migration.
--

CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

-- Differs from `river_queue` in that it tracks the queue state for a particular
-- active client.
CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

-- River main migration 006 [up]
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION river_job_state_in_bitmask(bitmask BIT(8), state river_job_state)
RETURNS boolean
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending'   THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running'   THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;
-- +goose StatementEnd

--
-- Add `river_job.unique_states` and bring up an index on it.
--
-- This column may exist already if users manually created the column and index
-- as instructed in the changelog so the index could be created `CONCURRENTLY`.
--
ALTER TABLE river_job ADD COLUMN IF NOT EXISTS unique_states BIT(8);

-- This statement uses `IF NOT EXISTS` to allow users with a `river_job` table
-- of non-trivial size to build the index `CONCURRENTLY` out of band of this
-- migration, then follow by completing the migration.
CREATE UNIQUE INDEX IF NOT EXISTS river_job_unique_idx ON river_job (unique_key)
    WHERE unique_key IS NOT NULL
      AND unique_states IS NOT NULL
      AND river_job_state_in_bitmask(unique_states, state);

-- Remove the old unique index. Users who are actively using the unique jobs
-- feature and who wish to avoid deploy downtime may want od drop this in a
-- subsequent migration once all jobs using the old unique system have been
-- completed (i.e. no more rows with non-null unique_key and null
-- unique_states).
DROP INDEX river_job_kind_unique_key_idx;

-- River main migration 007 [up]
--
-- Notification outbox.
--

CREATE TABLE river_notification (
    id bigserial PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    payload text NOT NULL,
    topic text NOT NULL,
    CONSTRAINT topic_length CHECK (length(topic) > 0 AND length(topic) < 128)
);

CREATE INDEX river_notification_created_at_idx ON river_notification (created_at);
CREATE INDEX river_notification_topic_id_idx ON river_notification (topic, id);

--
-- SQLite JSONB conversion.
--
-- No-op. Postgres already stores River JSON columns as jsonb.

--
-- SQL cleanup.
--

--
-- Drop unused tables `river_client` and `river_client_queue`.
--

DROP TABLE river_client_queue;
DROP TABLE river_client;

--
-- Adds `DEFAULT 25` to `river_job.max_attempts`.
--

ALTER TABLE river_job
    ALTER COLUMN max_attempts SET DEFAULT 25;

--
-- Changes `river_queue.updated_at` to have a default of `CURRENT_TIMESTAMP`.
--

ALTER TABLE river_queue
    ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- River main migration 008 [up]
-- No-op. Postgres sequences already prevent automatically generated job IDs
-- from being reused.
SELECT 1;
-- +goose Down

-- River v0.51.2 canonical rollback from river migrate-get --all --down.
-- River main migration 008 [down]
-- No-op. Postgres sequences already prevent automatically generated job IDs
-- from being reused.
SELECT 1;

-- River main migration 007 [down]
--
-- SQL cleanup rollback.
--

--
-- Add back unused tables `river_client` and `river_client_queue`.
--

CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

--
-- Revert addition of `DEFAULT 25` to `river_job.max_attempts`.
--

ALTER TABLE river_job
    ALTER COLUMN max_attempts DROP DEFAULT;

--
-- Changes `river_queue.updated_at` to revert the default of `CURRENT_TIMESTAMP`.
--

ALTER TABLE river_queue
    ALTER COLUMN updated_at DROP DEFAULT;

--
-- SQLite JSONB conversion rollback.
--
-- No-op. Postgres already stores River JSON columns as jsonb.

--
-- Notification outbox rollback.
--

DROP TABLE river_notification;

-- River main migration 006 [down]
--
-- Drop `river_job.unique_states` and its index.
--

DROP INDEX river_job_unique_idx;

ALTER TABLE river_job
    DROP COLUMN unique_states;

CREATE UNIQUE INDEX IF NOT EXISTS river_job_kind_unique_key_idx ON river_job (kind, unique_key) WHERE unique_key IS NOT NULL;

--
-- Drop `river_job_state_in_bitmask` function.
--
DROP FUNCTION river_job_state_in_bitmask;

-- River main migration 005 [down]
--
-- Revert to migration table based only on `(version)`.
--
-- If any non-main migrations are present, 005 is considered irreversible.
--

-- +goose StatementBegin
DO
$body$
BEGIN
    -- Tolerate users who may be using their own migration system rather than
    -- River's. If they are, they will have skipped version 001 containing
    -- `CREATE TABLE river_migration`, so this table won't exist.
    IF (SELECT to_regclass('river_migration') IS NOT NULL) THEN
        IF EXISTS (
            SELECT *
            FROM river_migration
            WHERE line <> 'main'
        ) THEN
            RAISE EXCEPTION 'Found non-main migration lines in the database; version 005 migration is irreversible because it would result in loss of migration information.';
        END IF;

        ALTER TABLE river_migration
            RENAME TO river_migration_old;

        CREATE TABLE river_migration(
            id bigserial PRIMARY KEY,
            created_at timestamptz NOT NULL DEFAULT NOW(),
            version bigint NOT NULL,
            CONSTRAINT version CHECK (version >= 1)
        );

        CREATE UNIQUE INDEX ON river_migration USING btree(version);

        INSERT INTO river_migration
            (created_at, version)
        SELECT created_at, version
        FROM river_migration_old;

        DROP TABLE river_migration_old;
    END IF;
END;
$body$
LANGUAGE 'plpgsql';
-- +goose StatementEnd

--
-- Drop `river_job.unique_key`.
--

ALTER TABLE river_job
    DROP COLUMN unique_key;

--
-- Drop `river_client` and derivative.
--

DROP TABLE river_client_queue;
DROP TABLE river_client;

-- River main migration 004 [down]
ALTER TABLE river_job ALTER COLUMN args DROP NOT NULL;

ALTER TABLE river_job ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE river_job ALTER COLUMN metadata DROP DEFAULT;

-- It is not possible to safely remove 'pending' from the river_job_state enum,
-- so leave it in place.

ALTER TABLE river_job DROP CONSTRAINT finalized_or_finalized_at_null;
ALTER TABLE river_job ADD CONSTRAINT finalized_or_finalized_at_null CHECK (
  (state IN ('cancelled', 'completed', 'discarded') AND finalized_at IS NOT NULL) OR finalized_at IS NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION river_job_notify()
  RETURNS TRIGGER
  AS $$
DECLARE
  payload json;
BEGIN
  IF NEW.state = 'available' THEN
    -- Notify will coalesce duplicate notifications within a transaction, so
    -- keep these payloads generalized:
    payload = json_build_object('queue', NEW.queue);
    PERFORM
      pg_notify('river_insert', payload::text);
  END IF;
  RETURN NULL;
END;
$$
LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER river_notify
  AFTER INSERT ON river_job
  FOR EACH ROW
  EXECUTE PROCEDURE river_job_notify();

DROP TABLE river_queue;

ALTER TABLE river_leader
    ALTER COLUMN name DROP DEFAULT,
    DROP CONSTRAINT name_length,
    ADD CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128);

-- River main migration 003 [down]
ALTER TABLE river_job
    ALTER COLUMN tags DROP NOT NULL,
    ALTER COLUMN tags DROP DEFAULT;

-- River main migration 002 [down]
DROP TABLE river_job;
DROP FUNCTION river_job_notify;
DROP TYPE river_job_state;

DROP TABLE river_leader;

-- River main migration 001 [down]
DROP TABLE river_migration;
DROP TABLE api_keys;
DROP TABLE identity_limiter_buckets;
DROP TABLE identity_limiter_control;
DROP TABLE identity_enrollment_receipts;
DROP TABLE platform_staff_auth_buckets;
DROP TABLE platform_staff_auth_control;
DROP TABLE platform_staff_sessions;
DROP TABLE platform_staff_credentials;
DROP TABLE audit_events;
DROP TABLE platform_staff;
DROP TABLE organization_memberships;
DROP TABLE organizations;
DROP TABLE auth_sessions;
DROP TABLE email_action_tokens;
DROP TABLE password_credentials;
DROP TABLE users;
DROP FUNCTION identity_limiter_capacity_guard();
DROP FUNCTION identity_email_action_capacity_guard();
DROP FUNCTION core_staff_session_guard();
DROP FUNCTION core_staff_validate();
DROP FUNCTION core_staff_credential_changed();
DROP FUNCTION core_staff_changed();
DROP FUNCTION core_staff_control_guard();
DROP FUNCTION core_staff_fence();
DROP FUNCTION core_staff_read_only();
DROP FUNCTION core_owner_no_truncate();
DROP FUNCTION core_owner_validate();
DROP FUNCTION core_check_owner(uuid);
DROP FUNCTION core_owner_fence();
DROP FUNCTION audit_events_append_only();
DROP FUNCTION audit_detail_valid(jsonb);
