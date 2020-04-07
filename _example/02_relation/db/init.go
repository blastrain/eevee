package db

import (
	"database/sql"
	"io/ioutil"
	"relation/entity"

	_ "github.com/go-sql-driver/mysql"
)

func initUserRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS users"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/users.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `users` (`id`, `name`, `sex`, `age`, `skill_id`, `skill_rank`, `group_id`, `world_id`, `field_id`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	value := &entity.User{
		Name:      "john",
		Sex:       "man",
		Age:       30,
		SkillID:   1,
		SkillRank: 10,
		GroupID:   1,
		WorldID:   1,
		FieldID:   1,
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.Name,
		value.Sex,
		value.Age,
		value.SkillID,
		value.SkillRank,
		value.GroupID,
		value.WorldID,
		value.FieldID,
	); err != nil {
		return err
	}
	return nil
}

func initUserFieldRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS user_fields"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/user_fields.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `user_fields` (`id`, `user_id`, `field_id`) VALUES (?, ?, ?)"
	value := &entity.UserField{
		ID:      1,
		UserID:  1,
		FieldID: 1,
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.UserID,
		value.FieldID,
	); err != nil {
		return err
	}
	return nil
}

func initFieldRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS fields"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/fields.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `fields` (`id`, `name`, `location_x`, `location_y`, `object_num`, `level`, `difficulty`) VALUES (?, ?, ?, ?, ?, ?, ?)"
	value := &entity.Field{
		ID:         1,
		Name:       "fieldA",
		LocationX:  2,
		LocationY:  3,
		ObjectNum:  10,
		Level:      20,
		Difficulty: 5,
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.Name,
		value.LocationX,
		value.LocationY,
		value.ObjectNum,
		value.Level,
		value.Difficulty,
	); err != nil {
		return err
	}
	return nil
}

func initWorldRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS worlds"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/worlds.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `worlds` (`id`, `name`) VALUES (?, ?)"
	value := &entity.World{
		ID:   1,
		Name: "worldA",
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.Name,
	); err != nil {
		return err
	}
	return nil
}

func initSkillRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS skills"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/skills.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `skills` (`id`, `skill_effect`) VALUES (?, ?)"
	value := &entity.Skill{
		ID:          1,
		SkillEffect: "fire",
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.SkillEffect,
	); err != nil {
		return err
	}
	return nil
}

func initGroupRecord() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
	if _, err := conn.Exec("DROP TABLE IF EXISTS `groups`"); err != nil {
		return err
	}
	defer conn.Close()
	sql, err := ioutil.ReadFile("schema/groups.sql")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(string(sql)); err != nil {
		return err
	}
	query := "INSERT INTO `groups` (`id`, `name`) VALUES (?, ?)"
	value := &entity.Group{
		ID:   1,
		Name: "groupA",
	}
	if _, err := conn.Exec(
		query,
		value.ID,
		value.Name,
	); err != nil {
		return err
	}
	return nil
}

func initDB() error {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/?parseTime=true")
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec("CREATE DATABASE IF NOT EXISTS eevee"); err != nil {
		return err
	}
	return nil
}

func init() {
	if err := initDB(); err != nil {
		panic(err)
	}
	if err := initUserRecord(); err != nil {
		panic(err)
	}
	if err := initUserFieldRecord(); err != nil {
		panic(err)
	}
	if err := initFieldRecord(); err != nil {
		panic(err)
	}
	if err := initWorldRecord(); err != nil {
		panic(err)
	}
	if err := initSkillRecord(); err != nil {
		panic(err)
	}
	if err := initGroupRecord(); err != nil {
		panic(err)
	}
}
