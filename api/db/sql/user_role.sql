-- name: SelectRole :one
select * from user_role
where user_id = ?;

-- name: InitRole :exec
insert or ignore into user_role (user_id)
  values (?);