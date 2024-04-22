-- name: CreateImageInformation :execresult
insert into
  informations (name, created)
values
  (?, ?);
