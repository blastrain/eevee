# eevee

Generate model, repository, dao sources for Go application

<img width="300px" height="238px" src="https://user-images.githubusercontent.com/209884/29392112-b844b88e-8336-11e7-8435-2e472301cf36.png"></img>

eevee はアプリケーション開発時に必要となる  
キャッシュやデータベースといったミドルウェアとの効率的なデータのやりとりや  
開発時に生じる冗長な作業を自動化するための仕組みを提供します。

データをいかに簡単かつ効率的に参照し書き込めるかということにフォーカスしているため、  
ルーティングなどの機能は提供していません。  
そのため、 [echo](https://echo.labstack.com) や [chi](https://github.com/go-chi/chi) や [goji](https://github.com/goji/goji) といったアプリケーションフレームワークと同時に利用することを想定しています。

[goa](https://github.com/goadesign/goa) が提供しているような APIリクエスト・レスポンス を自動生成する機能等も存在しますが、  
プロジェクトにあわせて導入するしないを判断することができます。  

eevee が提供する機能は主に次のようなものです。

- スキーマ駆動開発によるモデル・リポジトリ層の自動生成
- モデル間の依存関係の自動解決
- `Eager Loading` / `Lazy Loading` を利用した効率的なデータ参照
- テスト開発を支援する mock インスタンス作成機能
- モデルからJSON文字列への高速な変換
- API リクエスト・レスポンスの自動生成
- プラグインを用いた柔軟なカスタマイズ

# 使い方

まずは実際に eevee を利用することで何ができるようになるのかを見ていきます。  
ここで紹介しているコードは [_example/01_simple](https://github.com/knocknote/eevee/tree/master/_example/01_simple) 配下に置かれています。

## 1. eevee のインストール

```bash
$ go get go.knocknote.io/eevee/cmd/eevee
```

無事インストールできていれば、 `$GOPATH/bin/eevee` があるはずです。  
`eevee help` が実行できれば、インストールは完了です

## 2. 作業ディレクトリの作成

アプリケーション開発のための作業用ディレクトリを作成してください

## 3. go.mod ファイルの作成

いつものように `go.mod` ファイルを作成してください

```bash
$ go mod init simple
```

## 4. アプリケーションコードの作成

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

## 5. スキーマファイルの作成

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

## 6. eevee の実行

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

## 7. アプリケーションコードの書き換え

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

### 1. アプリケーション実行のための準備

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

### 2. 作成したデータベースに対するコネクションの作成

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

### 3. 作成操作

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

### 4. 読み出し操作

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

### 5. 更新操作

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


### 6. 削除操作

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

### 1. `name` 
クラス名を記載します。スキーマから生成した場合は、スキーマ名の単数形になります

### 2. `datastore`

`dao` でやりとりする外部のミドルウェアの種類を記述します。  
`.eevee.yml` で設定したデフォルトの `datastore` の名前が書き込まれます。  
何も設定しない場合は `db` になります。

### 3. `index`

スキーマに記述した `PRIMARY KEY` , `UNIQUE KEY` , `KEY` の設定を反映したものになります。  
基本的にこの部分を手動で編集する必要はありません

### 4. `members`

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

### 4.1 `member.extend`

`extend: true` として定義したメンバは、 `entity` のメンバ変数には現れず、 `model` のメンバ変数にのみ追加されることを意味します。  

`eevee` は自動生成時に `entity` や `model` といったパッケージを生成しますが、
`entity` にはシリアライズ対象のメンバ変数のみをもたせ、アプリケーションロジックを記述する上で必要な状態変数などは `model` のメンバ変数として定義することを推奨しています。  

### 4.2 `member.render`

`model` として定義されたオブジェクトを `JSON` などにエンコードする際のふるまいをカスタマイズするために使用します。
`render: false` として定義されたメンバは、エンコード・デコードの
対象になりません。  
また、 `render: inline` として定義された場合は、後述する `relation` が存在する場合に `relation` 先のオブジェクトをエンコードした結果をマージします。

```yaml
render:
  json: group
```

のように、レンダリングプロトコルごとにエンコード・デコード時の名前を変更することができます。 `json` や `msgpack` など、複数のプロトコルに対応したい場合はこの記法を利用してください。 ( 何も設定しない場合は、 `lowerCamelCase` が使用されます )

### 4.3 `member.relation`

スキーマ間の依存関係を定義するためのパラメータです。  
このパラメータを定義することによって、
依存先のインスタンスを取得するためのアクセサが自動生成されるようになります。

#### 4.3.1 `member.relation.to`

依存先のクラス名を書きます。ここでのクラス名とは、クラスファイルの `name` パラメータです。

#### 4.3.2 `member.relation.internal`

依存先のインスタンスを取得するために利用する自クラスのメンバーの名前を指定します。

#### 4.3.3 `member.relation.external`

依存先のインスタンスを取得するために利用する依存先クラスのメンバーの名前を指定します。

#### 4.3.4 `member.relation.custom`

クラス間の紐付けルールが複雑な場合など、 `internal` , `external` の枠にとらわれずに依存先のインスタンスを取得したい場合に利用します。
このパラメータと `internal`, `external` パラメータを併用することはできません。

#### 4.3.5 `member.relation.all`

依存先クラスの値をすべて取得したい場合に利用します。`internal` , `external` パラメータと併用することはできません。

その他にも、以下のパラメータが利用できます。

### 4.4 `member.desc`

そのメンバの役割を把握するためのドキュメントを記述します。  
このパラメータは APIドキュメントを自動生成する際に利用されます。

### 4.5 `member.example`

そのメンバがとる値の例を記述します。ここで指定した値は、APIドキュメントの自動生成に利用される他、テスト時のモックオブジェクト作成用データとしても利用されます。

## 5. `read_only`

`read_only: true` と書くと、そのクラスは読み込み専用と解釈され、  
CRUD のうち READ 操作を行う API のみ自動生成されます。

利用例としては、アプリケーション開発でよく行われる、
管理者側であらかじめ用意したデータセット(マスターデータ)に用いることができます。
アプリケーションの利用者側から API を通して変更できるデータでない場合は `read_only: true` を指定すると安全に開発することができます。

## `type` の書き方について

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


# スキーマ駆動開発による、モデル・リポジトリ層の自動生成

`eevee` はスキーマ駆動開発を前提としています。  
はじめにスキーマを定義してそれを読み込むことで **クラスファイル** を生成し、
必要であればテーブル間の依存関係などを書き足した上で Go のソースコードを自動生成します。  

こうしてデータを扱いやすくするAPIを多く持ったモデルと、  
そのモデルを取得するためのリポジトリレイヤーを自動生成することでアプリケーション開発を効率的に行うことができます。  

自動生成されたパッケージは大まかに以下の図のような依存関係を持ちます

![eevee_arch](https://user-images.githubusercontent.com/209884/77878544-7e1fe580-7293-11ea-8241-3f66f2a9cc9e.png)

図は一番左から右へ、または一番右から左へ矢印の向きに沿って見ます。  
アプリケーション固有のロジックから `eevee` の機能を利用する場合、 `repository` と `model` パッケージを利用します。  
つまり、アプリケーションロジックを記述する場合は、この2つのパッケージに注力すれば良いことになります。  
`repository` , `model` は裏で `dao` パッケージを利用します。 `dao` は `DataAccessObject` の略で、アプリケーション外部とデータをやりとりするためのローレベルなAPIを提供します。  
`dao` が外部とやりとりする場合、必ず `entity` パッケージを利用します。  
`entity` にはシリアライズ可能なデータ構造が定義されており、  
このデータ構造を通して `dao` が適切にシリアライズ・デシリアライズすることで外部とのやりとりを実現します。

データの読み込み方向に注目すると以下の図のようになります。

![eevee_read_arch](https://user-images.githubusercontent.com/209884/77878551-81b36c80-7293-11ea-9e1c-50dc48322ff4.png)

大きく、 1 と 2 の 2通りの読み込み方法があります。
まずアプリケーションからデータを取得したいと思った場合は、 `repository` を利用します。 `repository` が提供する `CRUD` API を通して `dao` を経由しつつデータを取得します。このとき、 `dao` が扱うデータ構造は `entity` のため、アプリケーションから利用しやすいように `model` に変換することも行います。  

もうひとつは、 `model` を通して行う読み込み操作です。  
後述しますが、 `model` にはリレーション関係にある別のテーブルデータを効率的に取得する機能があります。こういった機能を利用する場合は、 `model` が裏で `repository` を経由してデータを取得します。  

一方、データの書き込み方向に注目すると以下の図のようになります。

![eevee_write_arch](https://user-images.githubusercontent.com/209884/77878556-8415c680-7293-11ea-8f2d-6f0c26995d57.png)

こちらも 2 通りの方法があり、シンプルなのは `repository` を使ったものです。  
`repository` にはそのまま `CRUD` ができる API があるので、そちらを通して `Create` , `Update` , `Delete` を実行すれば、 `dao` を通して書き込み操作が反映されます。  
一方、 `model` を通して書き込むことも可能です。  
その場合は、モデルの内容を作成・または更新したいものに書き換えた後に `Save()` を呼ぶことで行うことができます。  
あわせて、 `Create` `Update` `Delete` といった直感的な API も用意しているので、用途によって使い分けることも可能です。

# モデル間の依存関係の自動解決

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

# `Eager Loading` / `Lazy Loading` を利用した効率的なデータ参照

前項で触れましたが、 `eevee` にはモデル間の依存関係を解決する機能があります。  
この機能を提供する上で大切にしたのは次の2点です。  

1. `N+1` 問題が起こらないこと
2. 不必要なデータを読まないこと

それぞれどのように解決しているのかを説明します。

## `Eager Loading` を用いた `N + 1` 問題の解決

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

## `Lazy Loading` を用いた効率的なデータ参照

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

## API リクエスト・レスポンスの自動生成

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

# Committers

- Masaaki Goshima ( [goccy](https://github.com/goccy) )


# License

MIT
