# How to work example

## 1. Automatically resolve relationships between tables at render

```go
func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo := repository.New(ctx, tx)
	user, err := repo.User().FindByID(ctx, uint64(id))
	if err != nil {
		return err
	}
	if err := c.JSON(http.StatusOK, user); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
```

```bash
$ curl localhost:1323/users/1 | jq '.'
```

```json
{
  "id": 1,
  "name": "john",
  "sex": "man",
  "age": 30,
  "skillId": 1,
  "skillRank": 10,
  "groupId": 1,
  "worldId": 1,
  "fieldId": 1,
  "userFields": [
    {
      "id": 1,
      "userId": 1,
      "fieldId": 1,
      "field": {
        "id": 1,
        "name": "fieldA",
        "locationX": 2,
        "locationY": 3,
        "objectNum": 10,
        "level": 20,
        "difficulty": 5
      }
    }
  ],
  "skillEffect": "fire",
  "group": null
}
```
