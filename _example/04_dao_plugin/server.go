package main

import (
	"context"
	_ "daoplugin/db"
	"daoplugin/repository"
	"database/sql"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	db *sql.DB
)

func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo := repository.New(ctx, tx, uint64(id))
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

func main() {
	e := echo.New()

	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	db = conn

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/users/:id", getUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
