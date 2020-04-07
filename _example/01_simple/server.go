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
