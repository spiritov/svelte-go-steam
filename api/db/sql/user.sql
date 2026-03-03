-- name: SelectUser :one
select * from user
where user.id = ?;

-- name: UpsertUser :one
insert into user (id, alias, avatar_url)
  values (?, ?, ?)
  on conflict do update
  set alias = excluded.alias,
  avatar_url = excluded.avatar_url
  returning *;