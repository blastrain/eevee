# eevee

[![GoDoc](https://godoc.org/go.knocknote.io/eevee?status.svg)](https://pkg.go.dev/mod/go.knocknote.io/eevee)
![Go](https://github.com/knocknote/eevee/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/knocknote/eevee/branch/master/graph/badge.svg)](https://codecov.io/gh/knocknote/eevee)
[![Go Report Card](https://goreportcard.com/badge/go.knocknote.io/eevee)](https://goreportcard.com/report/go.knocknote.io/eevee)

Generate model, repository, dao sources for Go application

<img width="500px" src="https://user-images.githubusercontent.com/209884/79103149-0ed3e680-7da7-11ea-8b89-474a01db7bad.png"></img>

`eevee` はアプリケーション開発時に必要となる  
キャッシュやデータベースといったミドルウェアとの効率的なデータのやりとりや  
開発時に生じる冗長な作業を自動化するための仕組みを提供します。

データをいかに簡単かつ効率的に参照し書き込めるかということにフォーカスしているため、  
ルーティングなどの機能は提供していません。  
そのため、 [echo](https://echo.labstack.com) や [chi](https://github.com/go-chi/chi) や [goji](https://github.com/goji/goji) といったアプリケーションフレームワークと同時に利用することを想定しています。

[goa](https://github.com/goadesign/goa) が提供しているような APIリクエスト・レスポンス を自動生成する機能等も存在しますが、  
プロジェクトにあわせて導入するしないを判断することができます。  

`eevee` が提供する機能は主に次のようなものです。

- スキーマ駆動開発によるモデル・リポジトリ層の自動生成
- モデル間の依存関係の自動解決
- `Eager Loading` / `Lazy Loading` を利用した効率的なデータ参照
- テスト開発を支援する mock インスタンス作成機能
- モデルからJSON文字列への高速な変換
- API リクエスト・レスポンスとそのドキュメントの自動生成
- プラグインを用いた柔軟なカスタマイズ

`eevee` は 600 を超えるテーブル、150万行を超える規模のアプリケーション開発を日々支えており、  
小規模開発から大規模開発まで様々な用途で利用することができます。

<!-- TOC -->

# 目次

- [使い方](#%E4%BD%BF%E3%81%84%E6%96%B9)
    - [eevee のインストール](#eevee-%E3%81%AE%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB)
    - [作業ディレクトリの作成](#%E4%BD%9C%E6%A5%AD%E3%83%87%E3%82%A3%E3%83%AC%E3%82%AF%E3%83%88%E3%83%AA%E3%81%AE%E4%BD%9C%E6%88%90)
    - [go.mod ファイルの作成](#gomod-%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%81%AE%E4%BD%9C%E6%88%90)
    - [アプリケーションコードの作成](#%E3%82%A2%E3%83%97%E3%83%AA%E3%82%B1%E3%83%BC%E3%82%B7%E3%83%A7%E3%83%B3%E3%82%B3%E3%83%BC%E3%83%89%E3%81%AE%E4%BD%9C%E6%88%90)
    - [スキーマファイルの作成](#%E3%82%B9%E3%82%AD%E3%83%BC%E3%83%9E%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%81%AE%E4%BD%9C%E6%88%90)
    - [eevee の実行](#eevee-%E3%81%AE%E5%AE%9F%E8%A1%8C)
    - [アプリケーションコードの書き換え](#%E3%82%A2%E3%83%97%E3%83%AA%E3%82%B1%E3%83%BC%E3%82%B7%E3%83%A7%E3%83%B3%E3%82%B3%E3%83%BC%E3%83%89%E3%81%AE%E6%9B%B8%E3%81%8D%E6%8F%9B%E3%81%88)
        - [アプリケーション実行のための準備](#%E3%82%A2%E3%83%97%E3%83%AA%E3%82%B1%E3%83%BC%E3%82%B7%E3%83%A7%E3%83%B3%E5%AE%9F%E8%A1%8C%E3%81%AE%E3%81%9F%E3%82%81%E3%81%AE%E6%BA%96%E5%82%99)
        - [作成したデータベースに対するコネクションの作成](#%E4%BD%9C%E6%88%90%E3%81%97%E3%81%9F%E3%83%87%E3%83%BC%E3%82%BF%E3%83%99%E3%83%BC%E3%82%B9%E3%81%AB%E5%AF%BE%E3%81%99%E3%82%8B%E3%82%B3%E3%83%8D%E3%82%AF%E3%82%B7%E3%83%A7%E3%83%B3%E3%81%AE%E4%BD%9C%E6%88%90)
        - [作成操作](#%E4%BD%9C%E6%88%90%E6%93%8D%E4%BD%9C)
        - [読み出し操作](#%E8%AA%AD%E3%81%BF%E5%87%BA%E3%81%97%E6%93%8D%E4%BD%9C)
        - [更新操作](#%E6%9B%B4%E6%96%B0%E6%93%8D%E4%BD%9C)
        - [削除操作](#%E5%89%8A%E9%99%A4%E6%93%8D%E4%BD%9C)
- [設定ファイル](#%E8%A8%AD%E5%AE%9A%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB)
    - [全体設定ファイル ( `.eevee.yml` )](#%E5%85%A8%E4%BD%93%E8%A8%AD%E5%AE%9A%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB--eeveeyml-)
        - [`module`](#module)
        - [`schema`](#schema)
        - [`class`](#class)
        - [`graph`](#graph)
        - [`output`](#output)
        - [`api`](#api)
        - [`document`](#document)
        - [`dao`](#dao)
            - [`dao.name`](#daoname)
            - [`dao.default`](#daodefault)
            - [`dao.datastore`](#daodatastore)
        - [`entity`](#entity)
            - [`entity.name`](#entityname)
            - [`entity.plugins`](#entityplugins)
        - [`model`](#model)
            - [`model.name`](#modelname)
        - [`repository`](#repository)
            - [`repository.name`](#repositoryname)
        - [`context`](#context)
            - [`context.import`](#contextimport)
        - [`plural`](#plural)
            - [`plural[].name`](#pluralname)
            - [`plural[].one`](#pluralone)
        - [`renderer`](#renderer)
            - [`renderer.style`](#rendererstyle)
        - [`primitivetypes`](#primitivetypes)
            - [`primitivetypes[].name`](#primitivetypesname)
            - [`primitivetypes[].packagename`](#primitivetypespackagename)
            - [`primitivetypes[].default`](#primitivetypesdefault)
            - [`primitivetypes[].as`](#primitivetypesas)
    - [クラスファイル](#%E3%82%AF%E3%83%A9%E3%82%B9%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB)
        - [`name`](#name)
        - [`datastore`](#datastore)
        - [`index`](#index)
        - [`members`](#members)
        - [`member.extend`](#memberextend)
        - [`member.render`](#memberrender)
        - [`member.relation`](#memberrelation)
            - [`member.relation.to`](#memberrelationto)
            - [`member.relation.internal`](#memberrelationinternal)
            - [`member.relation.external`](#memberrelationexternal)
            - [`member.relation.custom`](#memberrelationcustom)
            - [`member.relation.all`](#memberrelationall)
        - [`member.desc`](#memberdesc)
        - [`member.example`](#memberexample)
        - [`readonly`](#readonly)
        - [`type` の書き方について](#type-%E3%81%AE%E6%9B%B8%E3%81%8D%E6%96%B9%E3%81%AB%E3%81%A4%E3%81%84%E3%81%A6)
    - [API 定義ファイル](#api-%E5%AE%9A%E7%BE%A9%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB)
        - [`name`](#name)
        - [`desc`](#desc)
        - [`uri`](#uri)
        - [`method`](#method)
        - [`response`](#response)
            - [`response.type`](#responsetype)
            - [`response.subtypes`](#responsesubtypes)
            - [`response.type.include` ( `response.subtypes[].include` )](#responsetypeinclude--responsesubtypesinclude-)
            - [`include.name`](#includename)
            - [`include.only`](#includeonly)
            - [`include.except`](#includeexcept)
            - [`include.include`](#includeinclude)
        - [`includeall`](#includeall)
- [各機能について](#%E5%90%84%E6%A9%9F%E8%83%BD%E3%81%AB%E3%81%A4%E3%81%84%E3%81%A6)
    - [スキーマ駆動開発による、モデル・リポジトリ層の自動生成](#%E3%82%B9%E3%82%AD%E3%83%BC%E3%83%9E%E9%A7%86%E5%8B%95%E9%96%8B%E7%99%BA%E3%81%AB%E3%82%88%E3%82%8B%E3%83%A2%E3%83%87%E3%83%AB%E3%83%BB%E3%83%AA%E3%83%9D%E3%82%B8%E3%83%88%E3%83%AA%E5%B1%A4%E3%81%AE%E8%87%AA%E5%8B%95%E7%94%9F%E6%88%90)
    - [モデル間の依存関係の自動解決](#%E3%83%A2%E3%83%87%E3%83%AB%E9%96%93%E3%81%AE%E4%BE%9D%E5%AD%98%E9%96%A2%E4%BF%82%E3%81%AE%E8%87%AA%E5%8B%95%E8%A7%A3%E6%B1%BA)
    - [`Eager Loading` / `Lazy Loading` を利用した効率的なデータ参照](#eager-loading--lazy-loading-%E3%82%92%E5%88%A9%E7%94%A8%E3%81%97%E3%81%9F%E5%8A%B9%E7%8E%87%E7%9A%84%E3%81%AA%E3%83%87%E3%83%BC%E3%82%BF%E5%8F%82%E7%85%A7)
        - [`Eager Loading` を用いた `N + 1` 問題の解決](#eager-loading-%E3%82%92%E7%94%A8%E3%81%84%E3%81%9F-n--1-%E5%95%8F%E9%A1%8C%E3%81%AE%E8%A7%A3%E6%B1%BA)
        - [`Lazy Loading` を用いた効率的なデータ参照](#lazy-loading-%E3%82%92%E7%94%A8%E3%81%84%E3%81%9F%E5%8A%B9%E7%8E%87%E7%9A%84%E3%81%AA%E3%83%87%E3%83%BC%E3%82%BF%E5%8F%82%E7%85%A7)
    - [テスト開発を支援する mock インスタンス作成機能](#%E3%83%86%E3%82%B9%E3%83%88%E9%96%8B%E7%99%BA%E3%82%92%E6%94%AF%E6%8F%B4%E3%81%99%E3%82%8B-mock-%E3%82%A4%E3%83%B3%E3%82%B9%E3%82%BF%E3%83%B3%E3%82%B9%E4%BD%9C%E6%88%90%E6%A9%9F%E8%83%BD)
    - [モデルからJSON文字列への高速な変換](#%E3%83%A2%E3%83%87%E3%83%AB%E3%81%8B%E3%82%89json%E6%96%87%E5%AD%97%E5%88%97%E3%81%B8%E3%81%AE%E9%AB%98%E9%80%9F%E3%81%AA%E5%A4%89%E6%8F%9B)
    - [API リクエスト・レスポンスとそのドキュメントの自動生成](#api-%E3%83%AA%E3%82%AF%E3%82%A8%E3%82%B9%E3%83%88%E3%83%BB%E3%83%AC%E3%82%B9%E3%83%9D%E3%83%B3%E3%82%B9%E3%81%A8%E3%81%9D%E3%81%AE%E3%83%89%E3%82%AD%E3%83%A5%E3%83%A1%E3%83%B3%E3%83%88%E3%81%AE%E8%87%AA%E5%8B%95%E7%94%9F%E6%88%90)
    - [プラグインを用いた柔軟なカスタマイズ](#%E3%83%97%E3%83%A9%E3%82%B0%E3%82%A4%E3%83%B3%E3%82%92%E7%94%A8%E3%81%84%E3%81%9F%E6%9F%94%E8%BB%9F%E3%81%AA%E3%82%AB%E3%82%B9%E3%82%BF%E3%83%9E%E3%82%A4%E3%82%BA)
- [eevee による実践的な開発方法](#eevee-%E3%81%AB%E3%82%88%E3%82%8B%E5%AE%9F%E8%B7%B5%E7%9A%84%E3%81%AA%E9%96%8B%E7%99%BA%E6%96%B9%E6%B3%95)
    - [watch モードを利用する](#watch-%E3%83%A2%E3%83%BC%E3%83%89%E3%82%92%E5%88%A9%E7%94%A8%E3%81%99%E3%82%8B)
    - [repository に API を追加する](#repository-%E3%81%AB-api-%E3%82%92%E8%BF%BD%E5%8A%A0%E3%81%99%E3%82%8B)
    - [dao に実装されている一部の API の中身を自由に書き変える](#dao-%E3%81%AB%E5%AE%9F%E8%A3%85%E3%81%95%E3%82%8C%E3%81%A6%E3%81%84%E3%82%8B%E4%B8%80%E9%83%A8%E3%81%AE-api-%E3%81%AE%E4%B8%AD%E8%BA%AB%E3%82%92%E8%87%AA%E7%94%B1%E3%81%AB%E6%9B%B8%E3%81%8D%E5%A4%89%E3%81%88%E3%82%8B)
    - [model に API を追加する](#model-%E3%81%AB-api-%E3%82%92%E8%BF%BD%E5%8A%A0%E3%81%99%E3%82%8B)
    - [`relation.custom` を利用する](#relationcustom-%E3%82%92%E5%88%A9%E7%94%A8%E3%81%99%E3%82%8B)
- [Committers](#committers)
- [License](#license)

<!-- /TOC -->

# 使い方

まずは実際に eevee を利用することで何ができるようになるのかを見ていきます。  
ここで紹介しているコードは [_example/01_simple](https://github.com/knocknote/eevee/tree/master/_example/01_simple) 配下に置かれています。

## eevee のインストール

```bash
$ go get go.knocknote.io/eevee/cmd/eevee
```

無事インストールできていれば、 `$GOPATH/bin/eevee` があるはずです。  
`eevee help` が実行できれば、インストールは完了です

## 作業ディレクトリの作成

アプリケーション開発のための作業用ディレクトリを作成してください

## go.mod ファイルの作成

いつものように `go.mod` ファイルを作成してください

```bash
$ go mod init simple
```

## アプリケーションコードの作成

今回は [echo](https://echo.labstack.com) の https://echo.labstack.com/cookbook/crud をベースに
`eevee` を利用したいと思います。

リンク先にあるコードは以下のようになっています。

```go
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

このサンプルでは `user` というリソースに対して CRUD 操作を行っていますが、
サンプルのためデータはサーバのメモリ上に置かれています。  
これを eevee を用いて データベース(MySQL) 上への操作に変更することを行ってみます。

まずは、上記のコードを `server.go` として保存します。

## スキーマファイルの作成

`user` に関するデータを MySQL 上に保存することにしたので、まずはそのスキーマを定義します。  
次のようなコマンドで、 `id` と `name` というカラムをもった `users` テーブルを作るDDLが書かれたファイルを作成します。

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

## eevee の実行

eevee の `init` コマンドを実行します。

```bash
$ eevee init --schema schema --class config
```

`--schema` オプションでスキーマファイルが置かれているディレクトリを指定します ( 今回は `schema` )  
`--class` オプションでクラスファイルが生成されるディレクトリを指定します ( 今回は `config` )  

※ eevee で Go のソースコードを自動生成する際、上述の **クラスファイル** というものを参照します。  
これについては[後で詳しく](https://github.com/knocknote/eevee#%E3%82%AF%E3%83%A9%E3%82%B9%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB)説明します。

うまくいくと `.eevee.yml` というファイルが作成されているはずです。  
この状態で、以下のコマンドを実行してください

```bash
$ eevee run
```

先ほど作成した `.eevee.yml` を読み込み、定義に従ってソースコードの自動生成を行います。  
自動生成がうまくいくと、作業ディレクトリ配下は以下のようになるはずです。

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

新しく、 `config` `entity` `dao` `model` `repository` というディレクトリが作られています。

## アプリケーションコードの書き換え

それでは、自動生成されたコードを使って `server.go` を修正します。  
修正した後のコードは以下です。

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

ひとつずつ見ていきます。

### アプリケーション実行のための準備

```go
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
```

`func init()` で行っている処理は、このサンプルコードを動かすために
データベースや `users` テーブルを作成している処理です。  
ここでは、ローカルの MySQL サーバにつないで、 `eevee` という名前のデータベースを作成し、
そこに `users` テーブルを作成しています ( もしすでに存在してたら削除します )。

### 作成したデータベースに対するコネクションの作成

```go
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

`func main()` で行っている処理で変更があるのは

```go
conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/eevee?parseTime=true")
if err != nil {
  panic(err)
}

db = conn
```

の部分です。 `func init()` で作ったデータベースに対応するコネクションを作っています。
あわせて

```
var (
  db *sql.DB
)
```

で、 `db` インスタンスをグローバルに定義し、 `CRUD` 操作のいずれからも同じインスタンスを参照するようにします。

### 作成操作

CRUD のうち、 CREATE は以下のように変わりました。

```go
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
```

eevee は、あるリソースに対する CRUD 操作を `repository` パッケージを通して行うように設計されています。  
今回は `user` リソースを操作するため、 `repository.User` にアクセスするのですが、  
その方法は `repository.Repository` という共通のインスタンスを用いて行います。

共通インスタンスを作るためには、 `context.Context` と `*sql.Tx` が必要なので、それらを作ってから初期化します。

```go
ctx := context.Background()
tx, err := db.Begin()
...
repo := repository.New(ctx, tx)
```

次に `user` を作ります。 `repo.User()` で `user` リソースへアクセスすることができるようになり、  
リソース作成の場合は `Create(context.Context, *entity.User) (*model.User, error)` を実行します。

```go
user, err := repo.User().Create(ctx, reqUser)
if err != nil {
  return err
}
```

`entity.User` はロジックをもたない、シリアライズ可能なデータ構造です。  
まずは、リクエスト内容をこのインスタンスにマッピングすることで、作成したいデータ構造を表現します。

データの作成と同時に、 `*model.User` が返却されます。  
`model` パッケージにはアプリケーション開発に役立つ API が豊富に存在します ( 詳細は後述 )。  
ここではデータベース上に作成されたレコードと対応するインスタンスが返却されたと考えてください。

### 読み出し操作

CRUD のうち、 READ は以下のように変わりました。

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
  if err := tx.Commit(); err != nil {
    return err
  }
  return c.JSON(http.StatusOK, user)
}
```

まずは CREATE のときと同様に `repository.Repository` を作成することから行います。  
ここで、読み込み操作にも関わらず `db.Begin()` と `tx.Commit()` を実行していることが奇妙に思えるかもしれません。  
このような設計にしている背景として

1. CRUDのどの操作を行うかに関係なく、統一されたインターフェース ( `repository.New()` ) を通してリポジトリを作成し、操作してほしい  
2. 読み出しのみであってもキャッシュの作成など書き込み操作が含まれる場合もあり、トランザクションを前提としても良いケースもある(※ `repository.New()` には設定によって `sql.Tx` 以外のトランザクションインスタンスを渡すことができます )  

のようなものがあります。

ですが、それでも読み出しのみのAPIに関してはトランザクションを作りたくない場合があるかもしれません。   
そのため、現在の eevee では上記のような考え方に対する解は持っていませんが、  
声が多くある場合は、インターフェースを見直すことも検討しています。

```go
user, err := repo.User().FindByID(ctx, uint64(id))
if err != nil {
  return err
}
```

で `user` インスタンスを取得します。このとき得られるのは `*model.User` インスタンスです。  
モデルインスタンスは JSON 文字列への高速変換をサポートしているため、  
そのまま以下のように JSON として出力することができます。

```go
return c.JSON(http.StatusOK, user)
```

### 更新操作

CRUD のうち、 UPDATE は以下のように変わりました。

```go
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
```

`repository.Repository` を作成する流れは同じです。  
特筆すべきは、 `user.Save(ctx)` で更新処理を実現していることでしょう。

eevee が自動生成したコードを用いてリソースを更新するには、2通りの方法があります。  
一つは、 `repo.User().UpdateByID(context.Context, uint64, map[string]interface{}) error` を用いた方法です。  
`repository` パッケージを通して、更新したいレコードの ID と 更新したいカラム名とその値を map にしたものを渡します。  
この方法は直感的ではありますが `map[string]interface{}` を手動で作成する場合などは、  
値が正しいかをコンパイラで検査できないため、誤りを実行時にしか気づけないというデメリットもあります。  

もう一つは `Save(context.Context)` です。  
これは書き込み可能なモデルがもつ機能のうちのひとつで、  
モデルインスタンスをどういった手段で作ったかによって `Save(context.Context)` を呼んだ際に  
適切にリソースの作成または更新処理が走ります。

この `Save(context.Context` をリソースの作成・更新手段として利用することによって、  
レコードがあるとき・ないときといった場合分けや、どの値を更新すべきかといったことを意識する必要がありません。

今回のケースでは、 `FindByID()` で取得したインスタンスであることから、すでに存在するレコードを引いたことが自明なため、  
`Save()` を呼んだ際にレコードの更新処理が走ります。


### 削除操作

CRUD のうち、 DELETE は以下のように変わりました。

```go
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
```

`repository.Repository` を作って以下のように

```go
if err := repo.User().DeleteByID(ctx, uint64(id)); err != nil {
  return err
}
```

を呼べば、レコードを削除することができます。

ここまでで大まかな使い方の流れを理解してもらったところで、  
次は `eevee` が自動生成を行う際に参照しているファイルについて説明します。    

`eevee` を用いたアプリケーション開発を行う場合は、基本的にこれから説明する
二種類のファイルに設定を記述していきます。

# 設定ファイル

## 全体設定ファイル ( `.eevee.yml` )

`eevee init` をおこなうと、実行ディレクトリ配下にある `go.mod` を読み込んで
アプリケーション名を取得し、 `.eevee.yml` というファイルを作成します。  
`init` コマンド実行時にオプションを与えることで、あらかじめ設定値を `.eevee.yml` に反映させた上で生成することが可能ですが、あとから `YAML` ファイルを編集しても同じです。  

設定ファイルは一番簡素な状態だと以下のようなものです

`.eevee.yml`

```yaml
module: module_name
```

`module` には `go.mod` から読み込んだモジュール名が入ります ( もちろん変更可能です )。  
これでも良いのですが、もう少しここに書き足してみます。  
`eevee` はまずこの設定ファイルを読み込んで、後述する **クラスファイル** をどこに生成するかを判断します。  
デフォルトでは `eevee run` を実行したディレクトリ配下に生成するため、この挙動を変えたい場合は以下のように設定します。  

```yaml
module: module_name
class:  config
```

これで **クラスファイル** が `config` ディレクトリ配下に生成されるようになります。  
あわせて、 **クラスファイル** を生成するための元となるスキーマファイルを格納しているディレクトリも指定してみます。ここでは、 `schema` ディレクトリ配下にスキーマファイルを配置していることを前提として

```yaml
module: module_name
class:  config
schema: schema
```

と設定します。このように書くと、 `eevee` は `eevee run` を実行した際にスキーマファイルがどこにあるのかを把握して読み込み、その結果を **クラスファイル** として指定されたディレクトリ配下に自動生成し、さらにその内容から Go のソースコードを自動生成します。  

`.eevee.yml` には他にも多くのパラメータを設定することができ、自動生成全体に関わる挙動を変更することができます。

```yaml
module: module_name
graph:  graph
class:  config
schema: schema
plugins:
  plugin-name:
    repo: github.com/path/to/plugin-repo
dao:
  default: db
  datastore:
    db:
      before-create:
        - request-time
      before-update:
        - request-time
entity:
  plugins:
    - plugin-name
```

### `module`

Go のソースコードを自動生成する際に利用するモジュール名を指定します。
`go.mod` が存在する場合は自動的にモジュール名を読み取って書き込みます。

### `schema`

スキーマファイルを配置する場所を指定することができます

### `class`

クラスファイルを生成するパスを指定することができます

### `graph`

クラスファイルの依存関係を可視化したウェブページを表示する機能です。  
`graph: path/to/graph` と指定することで、指定先に `index.html` や `viz.js` といったページの表示に必要なファイルが生成されます。  

また、 `eevee serve` コマンドを実行すると、生成したファイルをサーブするウェブサーバが立ち上がります。

### `output`

`entity` `dao` `model` `repository` などのソースコードを自動生成する際の起点となるパスを指定することができます。デフォルトは `.` ( `eevee run` 実行時のディレクトリ ) です

### `api`

API定義が書かれた `YAML` ファイルが格納されているパスを指定することができます

### `document`

APIリクエスト・レスポンスを自動生成する機能を利用した際に、
同時に自動生成する APIドキュメント の生成場所を指定することができます

### `dao`

#### `dao.name`

`dao` というパッケージ名を変更するために使用します。
`dao` の役割はそのままに、名前だけを変更したい場合に利用します

#### `dao.default`

クラスファイルを自動生成する際に利用する、デフォルトの `datastore` を変更できます。　　
何も指定しない場合は `db` が使用されます

`datastore` にはリリース時点では `db` の他に `rapidash` が利用できます

#### `dao.datastore`

`datastore` の種類ごとにどのタイミングでどんなプラグインを使用して自動生成を行うかを指定することができます。  
以下の例では、 `db` では `create` と `update` 実行前のタイミングで、 `request-time` というプラグインを利用することを指定しています。同様に、 `datastore` として `rapidash` が指定された場合は `create` 実行前のタイミングで `other-plugin` というプラグインが使用されることを示しています。

```yaml
dao:
  datastore:
    db:
      before-create:
        - request-time
      before-update:
        - request-time
    rapidash:
      before-create:
        - other-plugin
```

### `entity`

#### `entity.name`

`entity` というパッケージ名を変更するために使用します。
`entity` の役割はそのままに、名前だけを変更したい場合に利用します

#### `entity.plugins`

`entity` のファイルを自動生成する際に使用するプラグインのリストを指定します

### `model`

#### `model.name`

`model` というパッケージ名を変更するために使用します。
`model` の役割はそのままに、名前だけを変更したい場合に利用します

### `repository`

#### `repository.name`

`repository` というパッケージ名を変更するために使用します。
`repository` の役割はそのままに、名前だけを変更したい場合に利用します

### `context`

`eevee` では、アプリケーション外部とやりとりする場面では常にAPIのインターフェースに `context.Context` を利用します。  
ですがアプリケーションによっては、独自の `context` パッケージを使用したい場合もあるでしょう。  
そういったケースに対応できるよう、 `context` パッケージのインポート先をカスタマイズすることができます。

#### `context.import`

デフォルトのインポート先 ( `context` ) を指定したパスでのインポートに置き換えます。  
インポート先のパッケージで `type Context interface {}` を定義し、 `context.Context` とコンパチのインターフェースを実装することで、アプリケーション独自の `context` に差し替えることができます。

### `plural`

自動生成コードに複数形の名前を用いたい場合、クラスファイルで指定された名前を利用して自動的に変換しています。  
ですが、すべての英単語を正しく変換できるわけではないため、
自動生成された名前が間違っている場合は、正しい名前を個別に指定する必要があります。  

#### `plural[].name`

複数形にした場合の名前を書きます

#### `plural[].one`

単数形の場合の名前を書きます

### `renderer`

#### `renderer.style`

クラスファイルのメンバごとにレンダリング時の挙動を指定するのが面倒な場合は、
このパラメータを指定することで一括で挙動を変更することができます。

- lower-camel : 出力時に lowerCamelCase を利用する
- upper-camel : 出力時に UpperCamelCase を利用する
- lower-snake : 出力時に lower_snake_case を利用する

### `primitive_types`

通常、クラスファイルのメンバには Go のプリミティブ型のみを指定するのですが、
アプリケーションによってはプリミティブ型を拡張した型を用いたい場合もあるでしょう。

(例)
```go
type ID uint64

func (id ID) MarshalJSON() ([]byte, error) {
  return []byte(fmt.Sprint(id)) // bigint をデコードできないクライアントのために、文字列として出力
}
```

上記のようなケースでは、以下のように `.eevee.yml` に記述しておくことでクラスファイル内で `ID` 型が利用できるようになります。

```yaml
primitive_types:
  - name: ID
    package_name: entity
    default: 1
    as: uint64
```

#### `primitive_types[].name`

型の名前を記述します

#### `primitive_types[].package_name`

対象の型がどのパッケージに属するかを指定します

#### `primitive_types[].default`

その型に値する値を出力する際のデフォルト値を指定します ( テストデータの自動生成に使用します )

#### `primitive_types[].as`

関連するプリミティブ型を指定します

## クラスファイル

クラスファイルは、eevee が Go のソースコードを自動生成する際に読み込むファイルです。  
基本的にはスキーマとクラスファイルは `1:1` の関係になります。  
ただし、クラスファイルを作るために必ずスキーマを書かないといけないわけではなく、
手でゼロから書くことも可能です。  
これは例えば、クラスファイルを作りたいが、そのクラスに対応するデータの保存先が RDBMS でない場合などに用います ( 例えば KVS に保存するなど )  

ここでは、説明のためにスキーマファイルがある前提で、  
そこからクラスファイルを自動生成する流れを説明します。

まず、例として以下のようなスキーマを持つテーブルからクラスファイルを生成します。

```sql
CREATE TABLE `users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) DEFAULT NULL,
  `sex` enum('man','woman') NOT NULL,
  `age` int NOT NULL,
  `skill_id` bigint(20) unsigned NOT NULL,
  `skill_rank` int NOT NULL,
  `group_id` bigint(20) unsigned NOT NULL,
  `world_id` bigint(20) unsigned NOT NULL,
  `field_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_users_01` (`name`),
  UNIQUE KEY `uq_users_02` (`skill_id`, `skill_rank`),
  KEY `idx_users_03` (`group_id`),
  KEY `idx_users_04` (`world_id`, `field_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

このスキーマを `schema/users.sql` として保存したあと、
`eevee init -s schema -c config` として `.eevee.yml` を生成し、 `eevee run` して作成された `config/user.yml` は以下のようになります。

```yaml
name: user
datastore: db
index:
  primary_key: id
  unique_keys:
  - - name
  - - skill_id
    - skill_rank
  keys:
  - - group_id
  - - world_id
    - field_id
members:
- name: id
  type: uint64
- name: name
  type: string
- name: sex
  type: string
- name: age
  type: int
- name: skill_id
  type: uint64
- name: skill_rank
  type: int
- name: group_id
  type: uint64
- name: world_id
  type: uint64
- name: field_id
  type: uint64
```

この内容をもとに、基本的なパラメータの説明をします。

### `name` 
クラス名を記載します。スキーマから生成した場合は、スキーマ名の単数形になります

### `datastore`

`dao` でやりとりする外部のミドルウェアの種類を記述します。  
クラスファイルを自動生成した場合、 `.eevee.yml` で設定したデフォルトの `datastore` が書き込まれます。 ( デフォルトは `db` です )
`datastore: none` とすると、 `eevee` で用意している自動生成コードが何も使用されない状態で `dao` のソースコードが生成されます ( 初回は空のファイルになります )。  
この機能は、 `eevee` の自動生成系には乗りたいが自分でファイルの内容をすべてカスタマイズしたい場合などに有効です。

### `index`

スキーマに記述した `PRIMARY KEY` , `UNIQUE KEY` , `KEY` の設定を反映したものになります。  
基本的にこの部分を手動で編集する必要はありません

### `members`

スキーマの各カラムに対応する定義を記述します。  
基本的には `name` と `type` の組み合わせとなり、 `name` はカラム名、 `type` は SQL での型を `Go` の型に変換したものが利用されます。  

以上がベースのパラメータになります。このままでも利用できますが、 `eevee` がもつテーブル間の参照解決機能を利用するために、いくつかのパラメータを追加します。  
以下の `YAML` 定義を参照してください。

```yaml
name: user
members:
... ( 省略 ) ...
- name: user_fields
  extend: true
  has_many: true
  relation:
    to: user_field
    internal: id
    external: user_id
- name: skill
  extend: true
  render:
    inline: true
  relation:
    to: skill
    internal: skill_id
    external: id
- name: group
  extend: true
  render:
    json: group
  relation:
    custom: true
    to: group
- name: world
  extend: true
  render: false
  relation:
    to: world
    internal: world_id
    external: id
```

`members` に新しく `user_fields` , `skill` , `group` , `world` を追加しました。  
それぞれの `member` で利用されているパラメータは以下のようなものです。

### `member.extend`

`extend: true` として定義したメンバは、 `entity` のメンバ変数には現れず、 `model` のメンバ変数にのみ追加されることを意味します。  

`eevee` は自動生成時に `entity` や `model` といったパッケージを生成しますが、
`entity` にはシリアライズ対象のメンバ変数のみをもたせ、アプリケーションロジックを記述する上で必要な状態変数などは `model` のメンバ変数として定義することを推奨しています。  

### `member.render`

`model` として定義されたオブジェクトを `JSON` などにエンコードする際のふるまいをカスタマイズするために使用します。  
`member.render` には複数の書き方が存在します。

1. `render: name` : 指定した名前を `key` にして値を出力します
2. `render: false` : `false` を指定するとエンコード・デコードの対象になりません
3. `render: inline` : `inline` を指定すると、後述する `relation` が存在する場合に `relation` 先のオブジェクトをエンコードした結果 ( key: value のペア )を自身の出力結果にマージします
4.
```yaml
render:
  json: lowerCamelName
  yaml: lower_snake_name
```

のように、レンダリングプロトコルごとにエンコード・デコード時の名前を変更することができます。 `json` や `msgpack` など、複数のプロトコルに対応したい場合はこの記法を利用してください。 ( 何も設定しない場合は、 `lowerCamelCase` が使用されます )

### `member.relation`

スキーマ間の依存関係を定義するためのパラメータです。  
このパラメータを定義することによって、
依存先のインスタンスを取得するためのアクセサが自動生成されるようになります。

#### `member.relation.to`

依存先のクラス名を書きます。ここでのクラス名とは、クラスファイルの `name` パラメータです。

#### `member.relation.internal`

依存先のインスタンスを取得するために利用する自クラスのメンバーの名前を指定します。

#### `member.relation.external`

依存先のインスタンスを取得するために利用する依存先クラスのメンバーの名前を指定します。

#### `member.relation.custom`

クラス間の紐付けルールが複雑な場合など、 `internal` , `external` の枠にとらわれずに依存先のインスタンスを取得したい場合に利用します。
このパラメータと `internal`, `external` パラメータを併用することはできません。

#### `member.relation.all`

依存先クラスの値をすべて取得したい場合に利用します。`internal` , `external` パラメータと併用することはできません。

その他にも、以下のパラメータが利用できます。

### `member.desc`

そのメンバの役割を把握するためのドキュメントを記述します。  
このパラメータは APIドキュメントを自動生成する際に利用されます。

### `member.example`

そのメンバがとる値の例を記述します。ここで指定した値は、APIドキュメントの自動生成に利用される他、テスト時のモックオブジェクト作成用データとしても利用されます。

### `read_only`

`read_only: true` と書くと、そのクラスは読み込み専用と解釈され、  
CRUD のうち READ 操作を行う API のみ自動生成されます。

利用例としては、アプリケーション開発でよく行われる、
管理者側であらかじめ用意したデータセット(マスターデータ)に用いることができます。
アプリケーションの利用者側から API を通して変更できるデータでない場合は `read_only: true` を指定すると安全に開発することができます。

### `type` の書き方について

`member.type` は複数の記述方法があります。  
もっともシンプルなのは、型をそのまま書く方法です。例えば以下のように書くことができます。

```yaml
members:
  - name: a
    type: interface{}
  - name: b
    type: map[string]interface{}
  - name: c
    type: *time.Time
```

しかし、上記の `time` パッケージは `Go` 標準の `time` パッケージでしょうか。  
アプリケーションによっては、別の `time` パッケージを参照したいかもしれません。  
そこで `eevee` では次のような記述方法もサポートしています。

```yaml
members:
  - name: d
    type:
      import: time
      package_name: time
      name: Time
      is_pointer: true
```

このように書くと、パッケージのインポートパスは `time` でパッケージの名前は `time` 、型名は `Time` でポインタであるということを意味します。  
この記法はアプリケーションで定義した型情報やサードパーティ製のライブラリで
利用されている構造体を定義したい場合に有効です。

## API 定義ファイル

API リクエストのパラメータを Go の構造体へマッピングする処理や、レスポンスに用いる JSON 文字列の作成支援を行う機能を利用するための設定方法について説明します。

はじめに、API定義は `YAML` を用いて記述しますが、そのファイルを格納するディレクトリを
`eevee` に教えるために、 `.eevee.yml` に次のように指定します。

`.eevee.yml`
```yaml
api: config/api
```

上記のように設定すると、 `config/api` 配下の `YAML` ファイルを読みに行き、
自動生成が走るようになります。

次に、ユーザー情報を返す API を例に `YAML` の書き方について説明します。
APIの名前を `user_getter` とし、次のように定義しました。

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
      members:
        - name: users
          type: user
          has_many: true
        - name: sub
          type: user_getter_subtype
      include:
        - name: user
          only:
            - id
            - name
          include:
            - name: user_fields
              only:
                - field_id
```

説明のために、あえて複雑なレスポンスを定義してみました。  
大事なのは `response` の部分で、ここに出力したいレスポンスの構造を記述していきます。  
APIはリストで記述します。つまり、1つのファイルに複数のAPI定義を書くことができます。  

### `name`

API名を記述します。この名前を利用してリクエストやレスポンスを処理するための構造体を作成します

### `desc`

API の説明を記述します。ドキュメントに反映されます

### `uri`

API にアクセスするための URI を記述します。ドキュメントに反映されます。

### `method`

HTTP メソッド ( `get` `post` `put` `delete` など ) を記載してください。  
指定したメソッドにあわせて、リクエストパラメータのデコード処理が変化します。

### `response`

`response` には `subtypes` と `type` を記述することができます

#### `response.type`
 
レスポンス用の構造体の定義を記述します。  
記述方法は、クラスファイルと同じように、 `members` を定義して行います。

#### `response.subtypes`

`response.type` を表現する際に階層構造を作りたい場合や、  
他のAPIとレスポンス構造をシェアしたい場合など、より複雑な構造を記述したい場合に利用します。  
記述方法は、クラスファイルと同じように、 `members` を定義して行います。

`subtypes` にはリスト構造で複数の `subtype` を定義することができます。


#### `response.type.include` ( `response.subtypes[].include` )

`type` , `subtype` には、 `members` の他に `include` プロパティがあります。  
`include` を適切に利用することで、依存関係にあるすべてのクラスを取得・レンダリングせずに、
必要な部分だけをレスポンスに含めることができます。  

#### `include.name`

レスポンスに含めたいメンバのうち

1. クラスファイルに定義されているもの
2. subtype として定義されているもの

の名前を記載します。

#### `include.only`

`include.name` で指定された定義のうち、 `members` の中でレスポンスに含めたいものを指定します。  
ここで指定できるのは、リレーション定義のないメンバのみです。

※ `include.except` と併用することはできません

#### `include.except`

`include.name` で指定された定義のうち、 `members` の中でレスポンスに含めたくないものを指定します。  
ここで指定できるのは、リレーション定義のないメンバのみです。

※ `include.only` と併用することはできません

#### `include.include`

`include.name` で指定された定義のうち、 `members` の中でリレーション定義をもつメンバに対して、
レスポンスに含めたいものの定義を記述します。  
再帰的に記述していくことが可能です。

### `include_all`

`include_all: true` と指定すると、すべての依存先メンバを含めます。  
( ただし、クラスファイル側で `render: false` が指定されている場合は出力されません )

# 各機能について

## スキーマ駆動開発による、モデル・リポジトリ層の自動生成

`eevee` はスキーマ駆動開発を前提としています。  
はじめにスキーマを定義してそれを読み込むことで **クラスファイル** を生成し、
必要であればテーブル間の依存関係などを書き足した上で Go のソースコードを自動生成します。  

こうしてデータを扱いやすくするAPIを多く持ったモデルと、  
そのモデルを取得するためのリポジトリレイヤーを自動生成することでアプリケーション開発を効率的に行うことができます。  

自動生成されたパッケージは大まかに以下の図のような依存関係を持ちます

<img src="https://user-images.githubusercontent.com/209884/77878544-7e1fe580-7293-11ea-8241-3f66f2a9cc9e.png" width="500px"/>

図は一番左から右へ、または一番右から左へ矢印の向きに沿って見ます。  
ビジネスロジックから `eevee` の機能を利用する場合、 `repository` と `model` パッケージを利用します。  
`repository` , `model` は裏で `dao` パッケージを利用します。 `dao` は `DataAccessObject` の略で、アプリケーション外部のプロセスとデータをやりとりするためのローレベルなAPIを提供します。  
`dao` が外部とやりとりする場合、必ず `entity` パッケージを利用します。  
`entity` にはシリアライズ可能なデータ構造が定義されており、  
このデータ構造を通して `dao` が適切にシリアライズ・デシリアライズすることで外部とのやりとりを実現します。

データの読み込み方向に注目すると以下の図のようになります。

<img src="https://user-images.githubusercontent.com/209884/77878551-81b36c80-7293-11ea-9e1c-50dc48322ff4.png" width="500px"/>

大きく、 1 と 2 の 2通りの読み込み方法があります。
まずアプリケーションからデータを取得したいと思った場合は、 `repository` を利用します。 `repository` が提供する `CRUD` API を通して `dao` を経由しつつデータを取得します。このとき、 `dao` が扱うデータ構造は `entity` のため、アプリケーションから利用しやすいように `model` に変換することも行います。  

もうひとつは、 `model` を通して行う読み込み操作です。  
後述しますが、 `model` にはリレーション関係にある別のテーブルデータを効率的に取得する機能があります。こういった機能を利用する場合は、 `model` が裏で `repository` を経由してデータを取得します。  

一方、データの書き込み方向に注目すると以下の図のようになります。

<img src="https://user-images.githubusercontent.com/209884/77878556-8415c680-7293-11ea-8f2d-6f0c26995d57.png" width="500px"/>

こちらも 2 通りの方法があり、シンプルなのは `repository` を使ったものです。  
`repository` にはそのまま `CRUD` ができる API があるので、そちらを通して `Create` , `Update` , `Delete` を実行すれば、 `dao` を通して書き込み操作が反映されます。  
一方、 `model` を通して書き込むことも可能です。  
その場合は、モデルの内容を作成・または更新したいものに書き換えた後に `Save()` を呼ぶことで行うことができます。  
あわせて、 `Create` `Update` `Delete` といった直感的な API も用意しているので、用途によって使い分けることも可能です。

## モデル間の依存関係の自動解決

**クラスファイル** にデータの依存関係を適切に記述することで開発を効率的に進めることができるようになります。  

例えば、 `users` テーブルの `id` カラムの値に対応する `user_id` というカラムを持った `user_fields` テーブルを考えてみます。　　
ここで、 `users` テーブルのレコードを取得してから `user_fields` のレコードを取得するには、通常次のように書くと思います。

```go
user, _ := repo.User().FindByID(ctx, 1)
userFields, _ := repo.UserField().FindByUserID(ctx, user.ID)
```

これを、両者の依存関係を **クラスファイル** に落とすと次のように書けます

```yaml
name: user
members:
... ( 省略 ) ...
- name: user_fields
  extend: true
  has_many: true
  relation:
    to: user_field
    internal: id
    external: user_id
```

上記は
1. `user` クラスは `user_field` クラスと依存関係にあり、
`user_fields` という名前のメンバでその依存関係が表現されている
2. `user` => `user_field` の参照は、 `user` クラスの `id` メンバと `user_field` クラスの `user_id` メンバの値を見て行われる
3. `user` => `user_field` の関係は `has_many` 関係にあるので、複数の `user_field` インスタンスを取得する
4. このメンバは `extend: true` がついているのでシリアライズ対象ではない ( `entity` には反映されない )

といったことを表しています。
この状態で `eevee run` を実行して Go のソースコードを自動生成すると、はじめに書いた `Go` のコードは次のように書くことができるようになります。

```go
user, _ := repo.User().FindByID(ctx, 1)
userFields, _ := user.UserFields(ctx)
```

この機能を用いることで、依存関係にあるクラスの値を簡単かつ安全に取得することができるようになります。  
またこのアクセスは非常に効率よく行われるので、例えば次のようなコレクションインスタンスに対して行われる際に特に有効です。 この例では通常 `N + 1` 回クエリが発行されてしまうように思われますが、実際には `Eager Loading` が行われ、2度のクエリ発行のみになります。このあたりの詳細は次の項目で説明します。

```go
user, _ := repo.User().FindByID(ctx, 1)
userFields, _ := user.UserFields(ctx)
userFields.Each(func(userField *model.UserField)) {
  // 普通はここで SELECT * FROM fields WHERE id = { user_field.field_id } のようなクエリが発行されるため効率が悪いが、
  // eevee では N+1 クエリを回避できる ( 後述 )
  field, _ := userField.Field(ctx)
}
```

この機能の重要な点は、あるクラスが関連するデータをすべてそのクラスのインスタンスから取得することができるということです。これによって、( エラー処理が入るので実際には利用感は異なりますが ) チェーンアクセスで依存データを取得することができたり、API レスポンスにあるインスタンスの関連データをすべて反映したりすることができるようになります。

## `Eager Loading` / `Lazy Loading` を利用した効率的なデータ参照

前項で触れましたが、 `eevee` にはモデル間の依存関係を解決する機能があります。  
この機能を提供する上で大切にしたのは次の2点です。  

1. `N+1` 問題が起こらないこと
2. 不必要なデータを読まないこと

それぞれどのように解決しているのかを説明します。

### `Eager Loading` を用いた `N + 1` 問題の解決

モデルをインスタンス化する際、 `eevee` では複数のインスタンスをまとめる場合、
スライスではなくコレクション構造体を利用します。
例えば複数の `user` インスタンスを取得する場合は `[]*model.User` ではなく、 `*model.Users` が返却されます

```go
users, _ := repo.User().FindByIDs(ctx, []uint64{1, 2, 3})
// users は []*model.User ではなく *model.Users
```

スライスではなく構造体を用いることで、  
このコレクションインスタンス自体がクエリ結果のキャッシュを持つことを可能にしています。  
例えば `_example/02_relation` を例にとると https://github.com/knocknote/eevee/blob/master/_example/02_relation/model/user.go#L44-L52 に書かれている通り `Users` 構造体として定義され、 `skills` や `userFields` といったキャッシュ用のメンバを持っていることが確認できると思います。

ではこの `skills` や `userFields` に値が入るのはいつかというと、 
例えば `skills` は `FindSkill` を呼んだときです。 ( つまり `user` がもつ `skillID` を使って `1:1` 対応する `skill` を引くタイミング )
https://github.com/knocknote/eevee/blob/master/_example/02_relation/model/user.go#L1396-L1409 

このメソッドの処理を読むと、 `skillID` を利用して `skill` を取得する際に、 `finder.FindByID(skillID)` とするのではなく、 `finder.FindByIDs(skillIDs)` と複数の `skillID` を使って取得しているのがわかると思います。  

つまり、 `*model.Users` はあらかじめ自身が管理する `*model.User` それぞれに対応する `skillID` の集合を `skillIDs` として保持しておき、どれかひとつでもその中の `skillID` を使って `skill` を検索する場合は、保持しておいた集合値を使ってすべての `skill` を取得しておき、その中から指定された `skillID` でフィルタするような挙動をとります。

これによって、毎回 `skillID` でクエリを投げることを防いでいるのがわかると思います。しかし、以下のような例ではコレクションインスタンスの `FindSkill()` が呼ばれるようなイメージが沸かないかもしれません。

```go
users, _ := repo.User().FindByIDs(ctx, []uint64{1, 2, 3})
users.Each(func(user *model.User)) {
  // *model.User から skill をとっているが、
  // 本当にコレクションインスタンスにアクセスしているのか
  skill, _ := user.Skill(ctx)
}
```

この答えは、 `repository` パッケージ内の次の箇所にあります。
https://github.com/knocknote/eevee/blob/master/_example/02_relation/repository/user.go#L411-L435


各 `*model.User` の他インスタンスを取得するためのアクセサは関数オブジェクトになっており、それをこの部分で作っています。  
このとき、コレクションインスタンスの参照(とそれを使ったメソッドコール)をクロージャを使って閉じ込めているため、各 `*model.User` インスタンスが `*model.Users` のメソッドを呼び出すことが可能になっているのです。

なので上記のコードは実は次のような処理になっています。

```go
users, _ := repo.User().FindByIDs(ctx, []uint64{1, 2, 3})
users.Each(func(user *model.User)) {
  // 1度目のアクセス時に SELECT * FROM skills WHERE id IN (1, 2, 3) 相当のことを行う
  // 取得した結果を users に保持しておく
  // users が保持している取得結果から該当の skill を検索する
  skill, _ := user.Skill(ctx)
}
```

### `Lazy Loading` を用いた効率的なデータ参照

前項で、他インスタンスへのアクセサが関数オブジェクトになっていることを説明しました。  
そのため、データを取得しにいくのは関数を読んときだけになります。  
あるインスタンスに紐づくデータを一番最初にすべて引きに行ってしまうと、無駄なデータを多く引いてしまう可能性がありますが、 `eevee` ではそのようなことはありません。  
必要なときに、必要なぶんだけデータを参照することができます。

## テスト開発を支援する mock インスタンス作成機能

`eevee` は `model` や `repository` といったパッケージを自動生成するのと同時に、 `mock/repository` と `mock/model/factory` というパッケージも生成します。  
これらはテスト開発時に `repository` 層をモックすることを支援してくれます。  

アプリケーション開発におけるテスト手法についてここで多くは触れませんが、  
データベースなどのミドルウェアにアクセスするようなテストケースを書く際、実際にアクセスして検証する方法と、モックを利用して擬似アクセスを行う方法があると思います。  

`eevee` では 後者のモックを用いた開発を支援しており、次のように `repository.Repository` インターフェースを実装した `*repository.RepositoryMock` インスタンスを返すことで、 `repository` 層のモックを簡単に行えるようにしています。

あわせて、 `model` インスタンスのファクトリパッケージも提供しており、簡単に `repository` の API を置き換え可能です。

```go
import (
  "app/mock/repository"
  "app/mock/model/factory"
)
repo := repository.NewMock()
ctx := context.Background()
repo.UserMock().EXPECT().FindByID(ctx, 1).Return(factory.DefaultUser(), nil)
```

mock 時のインターフェースは https://github.com/golang/mock を参考にしています。 `gomock` と違い型安全なコードを自動生成しているので、例えば上記の `FindByID` に与える引数の型に合わせるために `1` を `int64(1)` などとする必要はありません。 `gomock` は引数を `interface{}` で受けているためランタイム中に型エラーが発生する場合がありますが ( しかもわかりづらい )、そういった心配はありません。

`factory.DefaultUser()` の中身は、 `testdata/seeds` 配下の `YAML` ファイルを読み込んで自動生成しています。  

基本的にはインスタンスの初期値を `YAML` ファイルに名前付きで書いておくと、
その名前で `factory.XXX` という形でモデル初期化用の API を用意してくれるので、それを使います。

## モデルからJSON文字列への高速な変換

`model` パッケージを自動生成しているという設計上のメリットを使って、
`JSON` 文字列へエンコードする処理を静的に生成しています。これによって `reflect` 要らずとなっているため、 `encoding/json` を利用した `JSON` 文字列への変換よりも大分高速になっています。

## API リクエスト・レスポンスとそのドキュメントの自動生成

なぜ API リクエスト・レスポンスの作成までも支援しているか。  
それは `eevee` がもつメリットを最大限活かすために、レスポンス作成まで支援する必要があったからです。リクエストの方はレスポンス開発と同じ仕組みで開発できた方が便利だろうという理由からサポートしています。

前に説明していますが、`eevee` が自動生成するモデルインスタンスは、リレーション関係にある依存先のインスタンスも数珠つなぎでとってくることができます。  

モデルはそれぞれ `MarshalJSON` を実装しており、 `json.Marshal(user)` といったようにそのまま引数に与えると `JSON` 文字列に変換できます。  

ここで、デフォルトでは `user` に紐づく依存先のインスタンスも全て対象にして `JSON` 文字列へ変換しようとします。  

こうすることで、例えば `user` に紐づくモデルのどこかに依存関係を追加した場合 ( クラスファイルにメンバを追加する )、Go のコードを一切変更することなく `JSON` に結果を反映させることができます。

一度この仕組みを利用すると、これがどれだけ開発を楽にしているかを実感できると思いますが、メリットばかりではありません。  
すべての依存先インスタンスを取得、エンコードしようとするということは、それだけ時間もかかるし `JSON` のデータ量も大きくなります。  
API によっては、依存先の特定の部分だけ必要ないという場合もあるでしょう。  

そこで、 `eevee` では、各モデルに `ToJSON` と `ToJSONWithOption` という `JSON` 文字列化のための 2種類の手段を用意しています。  
`ToJSON` は `MarshalJSON` の内側で呼ばれる API で、依存先をすべて含めた `JSON` を作成しようと試みます。  
一方 `ToJSONWithOption` は、引数で与えたオプションによって、自身や依存先メンバの取捨選択ができるようになっています。この機能を用いることによって、「この API ではこのメンバだけ返す」といったことが容易にできるようになります。  
とはいえ、このオプションを API 作成のたびに `Go` で記述するのは大変です。  

そこで、 `Swagger` などの `YAML` を用いた API 開発を参考に、  
API レスポンスに必要なメンバの取捨選択もできるようにした上で `YAML` で記述できるようにし、それを読み込んで レスポンス作成に必要なソースコードを自動生成する機能をサポートしています。  

## プラグインを用いた柔軟なカスタマイズ

`eevee` が自動生成するパッケージ ( `entity` , `dao` , `repository` , `model` ) のうち、アプリケーションごとに設定が必要な部分は主に `dao` の部分です。

# eevee による実践的な開発方法

## watch モードを利用する

`eevee` を用いて数百を超えるクラスファイルを記述していくと、  
徐々に `eevee run` の時間が気になるようになっていきます。  

また、 クラスファイルや `dao` のソースコードを修正した場合に、
`eevee run` を実行するのを忘れてしまい、
自動生成対象を更新せずにコミットしてしまうようなことも起きるかもしれません。  

こういった問題を解決するために、
`eevee` には `-w` オプションをつけて起動することで、
ファイル変更イベントを監視して変更があったファイルに関連するファイルだけ
自動生成を走らせる機能 ( `watch` モード ) があります。  

1. クラスファイルに依存先の定義を適切に記述
2. そのクラスをレスポンスに含める
3. `eevee -w` ( `watch` モードでファイル監視 )

の3つを組み合わせることで、カラム追加など既存のデータ構造が変わるような場面で 「 クラスファイル ( `YAML` ファイル ) を変更した瞬間にレスポンス内容が変わる」という開発体験を得ることができます。  
この体験は今までのアプリケーション開発を変えるほどに良いため、 `watch` モードの利用を強く勧めています。

## repository に API を追加する

`repository` を用いてアクセスできる API は、スキーマファイルで定義したインデックス情報に基づいています。 ( `PRIMARY KEY` や `UNIQUE KEY` , `KEY` を適切に設定することで、それらのインデックスを用いた API を自動生成し、 `repository` を用いてアクセスできるようになります )  

このため、まずはインデックスを見直して生成するAPIを調整していただきたいですが、
自動生成対象になっているのは `Equal` で比較できるものだけになっています。  
それでは範囲検索や複雑なクエリを発行したい場合に困るので、
`eevee` には好きな API を `repository` に追加できる機能も存在します。  

`repository` に存在する API は、 `dao` に存在する公開 API をもとに自動生成しています。
( `repository` は完全自動生成のパッケージです。基本的に手動で何か処理を書き足すことはありません )  

例えば、以下のような `dao` パッケージのファイルがある場合 ( `_example/01_simple/dao/user.go` )

```go
package dao

import (
...
)

type User interface {
	Count(context.Context) (int64, error)
	Create(context.Context, *entity.User) error
	Delete(context.Context, *entity.User) error
	DeleteByID(context.Context, uint64) error
	DeleteByIDs(context.Context, []uint64) error
	FindAll(context.Context) (entity.Users, error)
	FindByID(context.Context, uint64) (*entity.User, error)
	FindByIDs(context.Context, []uint64) (entity.Users, error)
	Update(context.Context, *entity.User) error
	UpdateByID(context.Context, uint64, map[string]interface{}) error
	UpdateByIDs(context.Context, []uint64, map[string]interface{}) error
}

( 実装は省略 )
```

`repository` パッケージは次のようになります。

```go
// Code generated by eevee. DO NOT EDIT!

package repository

import (
...
)

type User interface {
	ToModel(*entity.User) *model.User
	ToModels(entity.Users) *model.Users
	Create(context.Context, *entity.User) (*model.User, error)
	Creates(context.Context, entity.Users) (*model.Users, error)
	FindAll(context.Context) (*model.Users, error)
	FindByID(context.Context, uint64) (*model.User, error)
	FindByIDs(context.Context, []uint64) (*model.Users, error)
	UpdateByID(context.Context, uint64, map[string]interface{}) error
	UpdateByIDs(context.Context, []uint64, map[string]interface{}) error
	DeleteByID(context.Context, uint64) error
	DeleteByIDs(context.Context, []uint64) error
	Count(context.Context) (int64, error)
	Delete(context.Context, *entity.User) error
	Update(context.Context, *entity.User) error
}

( 実装は省略 )
```

つまり、 `dao` に定義されている同名のクラスの `interface` の内容を解析して、 
`entity` の返り値を `model` のものに変換した内容が `repository` の `interface` に反映されます。  
このルールを利用すると、何か API を追加したい場合は以下のような手順で行うことができます。

1. `dao` に API を追加する

```go
package dao

import (
...
)

type User interface {
 ( 省略 )
 FindByRange(startAt time.Time, endAt time.Time) (entity.Users, error)
}

func (*UserImpl) FindByRange(startAt time.Time, endAt time.Time) (entity.Users, error) {
  ....
}
```

2. `eevee run` を実行

3. `repository.User` に API が追加される

```go

package repository

import (
  ...
)

type User interface {
  ( 省略 )
  FindByRange(startAt time.Time, endAt time.Time) (*model.Users, error)
}
```

## dao に実装されている一部の API の中身を自由に書き変える

「repository に API を追加する」で述べた機能を実現するため、
`dao` パッケージのファイルは同一ファイル内で自動生成と手動編集が混在することを許容しており、
**自動生成マーカー** によって実現しています。  

`dao` で自動生成されたAPIは、API毎に `// generated by eevee` というコメントが付与されています。  
このコメントは、対象APIが `eevee` によって自動生成されたものなのかを見分けるために使用されており、コメントがない API には自動生成の仕組みが適応されないようになっています。  

このため、自動生成された処理そのものを修正したい場合や自作のAPIなどは、
コメントが付いていない状況にしていただければ `eevee` 側で上書きなどはしません。

## model に API を追加する

`model` の構造体自体にメンバ変数を追加したい場合は、
クラスファイルの説明の中で触れた `extend` パラメータを利用してください。

```yaml
extend: true
```

を追加することで、モデルだけに任意のメンバ変数を追加することができます。

それとは別に、レシーバメソッドを追加したい場合もあるかと思います。  
そういった場合は、自動生成されたファイルとは別のファイルで ( 例えば `model/user_api.go` など ) 
以下のように好きな API を追加してください。

`model/user_api.go`
```go
package model

func (u *User) Hoge() {
  ...
}
```

## `relation.custom` を利用する

クラス間の依存関係を解決する際に、依存するパラメータの同値性以外で判断したいケースもあります。  
モデルに付与した状態変数によって A と B のクラスを切り替えたい場合もあるかもしれません。  
そういった場合は、依存解決を自分で実装できる `relation.custom` を用います。  

```yaml
relation:
  to: user_field
  custom: true
```

などと書けば、 「`UserField` クラスを参照するためのメンバ変数だが、依存解決方法は自作する」
といった意味になります。  

自動生成する際は

```go
func (u *User) UserField(ctx context.Context) (*UserField, error) {
  ...
}
```

**上記のような API が実装されることを前提として** 自動生成コードを生成します。  
つまり、上記の API を実装しなければコンパイルエラーになります。  

そのため、実装し忘れを防ぎつつ、 `UserField` を返すための処理を自作することができます。  

もうひとつ `relation.custom` の利用方法例として、ショートカットの実装があります。  

例えば A => B => C というクラスの依存関係がある場合、
A のインスタンスである `a` から C のインスタンスを取得する流れは次のようになります。

```go
b, _ := a.B(ctx)
c, _ := b.C(ctx)
```

これを

```go
c, _ := a.C(ctx)
```

と書けるようにするのがここでのショートカットです。  

実装方法は A のクラスファイルに以下のようなメンバを追加するだけです

```yaml
name: c
relation:
  to: c
  custom: true
```

こうすると、 `func (a *A) C(ctx context.Context) (*C, error)` という API が
実装されることを期待するので、実装します。

```go
func (a *A) C(ctx context.Context) (*C, error) {
  b, err := a.B(ctx)
  if err != nil {
    return nil, err
  }
  c, err := b.C(ctx)
  if err != nil {
    return nil, err
  }
  return c, nil
}
```

最初に示したコードを再利用できるように実装しただけですが、
A => C の取得を B を意識せずに行いたい場合は重宝します。

# Committers

- Masaaki Goshima ( [goccy](https://github.com/goccy) )


# License

MIT
