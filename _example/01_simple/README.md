# How to make application from scratch

## 1. Install eevee

```bash
$ go get go.knocknote.io/eevee/cmd/eevee
```

## 2. Create go.mod

```bash
$ go mod init simple
```

## 3. Add schema file

```bash
$ mkdir schema
$ cat <<'EOS' >> schema/users.sql
CREATE TABLE `users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
EOS
```

## 4. Execute eevee

```bash
$ eevee --schema schema --class config
```

generated the following files

```
├── config
│   └── user.yml
├── dao
│   └── user.go
├── entity
│   └── user.go
├── go.mod
├── go.sum
├── model
│   ├── model.go
│   └── user.go
├── repository
│   ├── repository.go
│   └── user.go
├── schema
   └── users.sql
```

## 5. Copy CRUD example from https://echo.labstack.com/cookbook/crud

```go : server.go
package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

var (
	users = map[int]*user{}
	seq   = 1
)

//----------
// Handlers
//----------

func createUser(c echo.Context) error {
	u := &user{
		ID: seq,
	}
	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = u
	seq++
	return c.JSON(http.StatusCreated, u)
}

func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/users", createUser)
	e.GET("/users/:id", getUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
```

## 6. Modify to use database by using generated files 

```go
package main

import (
	"context"
	"database/sql"
	"io/ioutil"
	"net/http"
	"simple/entity"
	"simple/repository"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	db *sql.DB
)

//----------
// Handlers
//----------

func createUser(c echo.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo := repository.New(ctx, tx)
	reqUser := new(entity.User)
	if err := c.Bind(reqUser); err != nil {
		return err
	}
	user, err := repo.User().Create(ctx, reqUser)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, user)
}

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
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func updateUser(c echo.Context) error {
	reqUser := new(entity.User)
	if err := c.Bind(reqUser); err != nil {
		return err
	}
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
	user.Name = reqUser.Name
	if err := user.Save(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func deleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo := repository.New(ctx, tx)
	if err := repo.User().DeleteByID(ctx, uint64(id)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func init() {
	{
		conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/?parseTime=true")
		if err != nil {
			panic(err)
		}
		if _, err := conn.Exec("CREATE DATABASE IF NOT EXISTS eevee"); err != nil {
			panic(err)
		}
	}
	{
		conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
		if _, err := conn.Exec("DROP TABLE IF EXISTS users"); err != nil {
			panic(err)
		}
		sql, err := ioutil.ReadFile("schema/users.sql")
		if err != nil {
			panic(err)
		}
		if _, err := conn.Exec(string(sql)); err != nil {
			panic(err)
		}
	}
}

func main() {
	e := echo.New()

	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if err != nil {
		panic(err)
	}

	db = conn

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/users", createUser)
	e.GET("/users/:id", getUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
```