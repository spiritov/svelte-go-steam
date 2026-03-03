-- name: InsertSession :one
insert into session (user_id, token_id)
  values (?, ?)
  returning *;