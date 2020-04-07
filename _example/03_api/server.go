package main

import (
	_ "api/db"
	"api/model"
	"api/repository"
	"api/request"
	"api/response"
	"context"
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
	req, err := new(request.UserGetterBuilder).SetUserID(uint64(id)).Build(c.Request())
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo := repository.New(ctx, tx)
	user, err := repo.User().FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	response, err := new(response.UserGetterBuilder).SetSub(&response.UserGetterSubtype{
		User:   user,
		Param1: "sub_param1",
		Param2: 100,
	}).SetUsers(new(model.Users).Add(user)).Build(ctx)
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
