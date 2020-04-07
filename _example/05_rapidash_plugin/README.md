# How to work example

## 1. Prepare memcached and MySQL server

e.g.)

```bash
$ memcached -d
```

```bash
$ mysql.server start
```

## 2. Run server

```
$ go run server.go
```

## 3. Get user data

```bash
$ curl localhost:1323/users/1 | jq '.'
```

```json
{
  "id": 1,
  "name": "john",
  "sex": "man",
  "age": 30,
  "skillID": 1,
  "skillRank": 10,
  "groupID": 1,
  "worldID": 1,
  "fieldID": 1,
  "createdAt": 1585820122,
  "updatedAt": 1585820122
}
```
