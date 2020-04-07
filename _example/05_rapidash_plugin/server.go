package main

import (
	"context"
	"database/sql"
	"net/http"
	_ "rapidashplugin/db"
	"rapidashplugin/entity"
	"rapidashplugin/repository"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.knocknote.io/rapidash"
)

var (
	db    *sql.DB
	cache *rapidash.Rapidash
)

func begin() (*rapidash.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	cacheTx, err := cache.Begin(tx)
	if err != nil {
		return nil, err
	}
	return cacheTx, nil
}

func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	tx, err := begin()
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

func main() {
	e := echo.New()

	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	db = conn
	c, err := rapidash.New(
		rapidash.LogEnabled(true),
		rapidash.ServerAddrs([]string{"localhost:11211"}),
	)
	if err != nil {
		panic(err)
	}
	if err := c.WarmUp(conn, new(entity.User).Struct(), false); err != nil {
		panic(err)
	}
	if err := c.WarmUp(conn, new(entity.UserField).Struct(), false); err != nil {
		panic(err)
	}
	if err := c.WarmUp(conn, new(entity.Field).Struct(), true); err != nil {
		panic(err)
	}
	if err := c.WarmUp(conn, new(entity.World).Struct(), true); err != nil {
		panic(err)
	}
	if err := c.WarmUp(conn, new(entity.Skill).Struct(), true); err != nil {
		panic(err)
	}
	cache = c

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/users/:id", getUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
