## return user status

```
get /users/{userId}
```

### Request Parameters

| Name | Type | In | Required | Description | Example |
| ---- | ---- | -- | -------- | ----------- | ------- |
| **userID** | uint64 | path | true |  | `<no value>` |
| **session** | string | header | false |  | `<no value>` |

### Response Parameters

<details>

| Name | Type | Description | Example |
| ---- | ---- | ----------- | ------- |
| **users[0].id** | uint64 |  | `0` |
| **users[0].name** | string |  | `` |
| **users[0].sex** | string |  | `` |
| **users[0].age** | int |  | `0` |
| **users[0].skillID** | uint64 |  | `0` |
| **users[0].skillRank** | int |  | `0` |
| **users[0].groupID** | uint64 |  | `0` |
| **users[0].worldID** | uint64 |  | `0` |
| **users[0].fieldID** | uint64 |  | `0` |
| **users[0].userFields[0].id** | uint64 |  | `0` |
| **users[0].userFields[0].userID** | uint64 |  | `0` |
| **users[0].userFields[0].fieldID** | uint64 |  | `0` |
| **users[0].userFields[0].field.id** | uint64 |  | `0` |
| **users[0].userFields[0].field.name** | string |  | `` |
| **users[0].userFields[0].field.locationX** | int |  | `0` |
| **users[0].userFields[0].field.locationY** | int |  | `0` |
| **users[0].userFields[0].field.objectNum** | int |  | `0` |
| **users[0].userFields[0].field.level** | int |  | `0` |
| **users[0].userFields[0].field.difficulty** | int |  | `0` |
| **users[0].skill.id** | uint64 |  | `0` |
| **users[0].skill.skillEffect** | string |  | `` |
| **users[0].group.id** | uint64 |  | `0` |
| **users[0].group.name** | string |  | `` |
| **sub.user.id** | uint64 |  | `0` |
| **sub.user.name** | string |  | `` |
| **sub.user.sex** | string |  | `` |
| **sub.user.age** | int |  | `0` |
| **sub.user.skillID** | uint64 |  | `0` |
| **sub.user.skillRank** | int |  | `0` |
| **sub.user.groupID** | uint64 |  | `0` |
| **sub.user.worldID** | uint64 |  | `0` |
| **sub.user.fieldID** | uint64 |  | `0` |
| **sub.user.userFields[0].id** | uint64 |  | `0` |
| **sub.user.userFields[0].userID** | uint64 |  | `0` |
| **sub.user.userFields[0].fieldID** | uint64 |  | `0` |
| **sub.user.userFields[0].field.id** | uint64 |  | `0` |
| **sub.user.userFields[0].field.name** | string |  | `` |
| **sub.user.userFields[0].field.locationX** | int |  | `0` |
| **sub.user.userFields[0].field.locationY** | int |  | `0` |
| **sub.user.userFields[0].field.objectNum** | int |  | `0` |
| **sub.user.userFields[0].field.level** | int |  | `0` |
| **sub.user.userFields[0].field.difficulty** | int |  | `0` |
| **sub.user.skill.id** | uint64 |  | `0` |
| **sub.user.skill.skillEffect** | string |  | `` |
| **sub.user.group.id** | uint64 |  | `0` |
| **sub.user.group.name** | string |  | `` |
| **sub.param1** | string |  | `` |
| **sub.param2** | int |  | `0` |

</details>


### Response Example

```json
{
    "sub": {
        "name": "",
        "param1": "",
        "param2": 0,
        "userFields": [
            {
                "field": {
                    "name": ""
                },
                "fieldID": 0
            }
        ]
    },
    "users": [
        {
            "id": 0,
            "name": "",
            "userFields": [
                {
                    "fieldID": 0
                }
            ]
        }
    ]
}
```
