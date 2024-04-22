-- name: CreateImageInformation :execresult
insert into
  informations (name, created)
values
  (?, ?);

-- name: ListInformations :many
SELECT
  *
FROM
  informations;

-- name: GetInformation :one
SELECT
  *
FROM
  informations
WHERE
  name = ?
LIMIT
  1;
