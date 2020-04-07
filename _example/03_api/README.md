# How to work example

## 1. Automatically filter response fields

- Define API definition to config/api/user.yml

```yaml
- name: user_getter
  desc: return user status
  uri: /users/:user_id
  method: get
  response:
    subtypes:
      - name: user_getter_subtype
        members:
          - name: user
            type: user
            render:
              inline: true
          - name: param1
            type: string
          - name: param2
            type: int
        include:
          - name: user
            only:
              - name
              - param1
              - param2
            include:
              - name: user_fields
                only:
                  - field_id
                include:
                  - name: field
                    only:
                      - name
    type:
      - name: user
        type: user_getter_subtype
        render:
          json: user
        desc: user status
      - name: param1
        type: string
        desc: param1
```

- Run eevee

```bash
$ eevee -r config/relation
```

- Use response package

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
	response, err := new(response.UserGetterBuilder).SetUser(&response.UserGetterSubtype{
		User:   user,
		Param1: "sub_param1",
		Param2: 100,
	}).SetParam1("param1").Build(ctx)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if err := c.JSON(http.StatusOK, response); err != nil {
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
  "user": {
    "name": "john",
    "userFields": [
      {
        "fieldId": 1,
        "field": {
          "name": "fieldA"
        }
      }
    ],
    "param1": "sub_param1",
    "param2": 100
  },
  "param1": "param1"
}
```
