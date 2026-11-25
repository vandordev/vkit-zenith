-- name: BOActivityRecords :many
WITH targets AS (
 SELECT 'staff'::text AS kind,id::text AS id,display_name AS name FROM platform_staff
 UNION ALL SELECT 'user',id::text,display_name FROM users
 UNION ALL SELECT 'organization',id::text,name FROM organizations
 UNION ALL SELECT 'organization_membership',m.id::text,u.display_name||' · '||o.name
 FROM organization_memberships m JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.organization_id
)
SELECT e.id,e.occurred_at,e.action,e.outcome,e.actor_kind,e.actor_staff_id,
 s.display_name AS actor_name,e.target_type,
 coalesce(CASE WHEN e.target_type='auth_bucket' THEN NULL::text ELSE e.target_id END,'')::text AS target_id,
 tm.organization_id AS target_organization_id,
 t.name AS target_name,(t.id IS NOT NULL
   AND (e.target_type<>'organization_membership' OR tm.id IS NOT NULL)
  AND (e.target_type<>'organization' OR (e.organization_id::text=e.target_id) IS TRUE)
  AND (e.target_type<>'user' OR e.organization_id IS NULL))::boolean AS target_available,
 CASE WHEN sqlc.arg(include_detail)::boolean AND octet_length(e.detail::text)<=65536
  THEN e.detail ELSE '{}'::jsonb END AS safe_detail
FROM audit_events e
LEFT JOIN platform_staff s ON s.id=e.actor_staff_id AND e.actor_kind='staff'
LEFT JOIN targets t ON t.kind=e.target_type AND t.id=e.target_id
LEFT JOIN organization_memberships tm ON e.target_type='organization_membership' AND tm.id::text=e.target_id AND tm.organization_id=e.organization_id
WHERE e.action=ANY(sqlc.arg(actions)::text[]) AND e.actor_kind IN ('staff','anonymous','system')
 AND (sqlc.narg(event_id)::uuid IS NULL OR e.id=sqlc.narg(event_id))
 AND (sqlc.narg(from_time)::timestamptz IS NULL OR e.occurred_at>=sqlc.narg(from_time))
 AND (sqlc.narg(until_time)::timestamptz IS NULL OR e.occurred_at<sqlc.narg(until_time))
 AND (sqlc.narg(actor_id)::uuid IS NULL OR e.actor_staff_id=sqlc.narg(actor_id))
 AND (sqlc.arg(action_filter)::text='' OR e.action=sqlc.arg(action_filter))
 AND (sqlc.arg(outcome_filter)::text='' OR e.outcome=sqlc.arg(outcome_filter))
ORDER BY e.occurred_at DESC,e.id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOActivityCount :one
WITH targets AS (
 SELECT 'staff'::text AS kind,id::text AS id,display_name AS name FROM platform_staff
 UNION ALL SELECT 'user',id::text,display_name FROM users
 UNION ALL SELECT 'organization',id::text,name FROM organizations
 UNION ALL SELECT 'organization_membership',m.id::text,u.display_name||' · '||o.name
 FROM organization_memberships m JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.organization_id
)
SELECT count(*) FROM audit_events e
LEFT JOIN platform_staff s ON s.id=e.actor_staff_id AND e.actor_kind='staff'
LEFT JOIN targets t ON t.kind=e.target_type AND t.id=e.target_id
WHERE e.action=ANY(sqlc.arg(actions)::text[]) AND e.actor_kind IN ('staff','anonymous','system')
 AND (sqlc.narg(event_id)::uuid IS NULL OR e.id=sqlc.narg(event_id))
 AND (sqlc.narg(from_time)::timestamptz IS NULL OR e.occurred_at>=sqlc.narg(from_time))
 AND (sqlc.narg(until_time)::timestamptz IS NULL OR e.occurred_at<sqlc.narg(until_time))
 AND (sqlc.narg(actor_id)::uuid IS NULL OR e.actor_staff_id=sqlc.narg(actor_id))
 AND (sqlc.arg(action_filter)::text='' OR e.action=sqlc.arg(action_filter))
 AND (sqlc.arg(outcome_filter)::text='' OR e.outcome=sqlc.arg(outcome_filter));

-- Choices are referenced actors (including disabled), searched by current name
-- only. selected_id uses this same restricted domain, not staff management.
-- name: BOActivityStaffChoices :many
SELECT s.id,s.display_name,s.status FROM platform_staff s
WHERE EXISTS (SELECT 1 FROM audit_events e WHERE e.actor_staff_id=s.id AND e.actor_kind='staff'
 AND e.action=ANY(sqlc.arg(actions)::text[]))
 AND (sqlc.arg(search)::text='' OR s.display_name ILIKE '%'||sqlc.arg(search)||'%')
 AND (sqlc.narg(selected_id)::uuid IS NULL OR s.id=sqlc.narg(selected_id))
ORDER BY s.display_name,s.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOActivityStaffChoiceCount :one
SELECT count(*) FROM platform_staff s
WHERE EXISTS (SELECT 1 FROM audit_events e WHERE e.actor_staff_id=s.id AND e.actor_kind='staff'
 AND e.action=ANY(sqlc.arg(actions)::text[]))
 AND (sqlc.arg(search)::text='' OR s.display_name ILIKE '%'||sqlc.arg(search)||'%');
