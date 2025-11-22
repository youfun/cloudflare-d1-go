# Cloudflare D1 Go 客户端 ☁️

[英文版本](README.md) | 中文

> 🔄 **增强版本说明**
> 
> 本项目是基于 [ashayas/cloudflare-d1-go](https://github.com/ashayas/cloudflare-d1-go) 的增强版本。
> 
> **主要改进：**
> - ✨ **sqlx 风格 API** - 清洁的方法如 `Select()`, `Get()`, `Exec()` 支持自动类型转换
> - ✨ **自动参数转换** - 直接传递 int、bool、time.Time，无需 []string 转换
> - ✨ **ConnectionPool 缓存** - 通过智能数据库连接池减少 99% 的 API 调用
> - ✨ 增强的数据类型处理（支持 D1 API 的数组格式行数据）
> - ✨ 高级查询支持（JOIN 查询、复杂 WHERE 条件）
> - ✨ 完整的 UPSERT 操作和 SQLite 冲突解决方案
> - ✨ 改进的错误处理和数据验证
> - ✨ StructScan 支持正确的 NULL 值处理
> - ✨ 真实场景的全面示例
>


<p align="center">
<a href="https://pkg.go.dev/github.com/youfun/cloudflare-d1-go"><img src="https://pkg.go.dev/badge/github.com/youfun/cloudflare-d1-go.svg" alt="Go Reference"></a>
<img src="https://img.shields.io/github/go-mod/go-version/youfun/cloudflare-d1-go" alt="Go Version">
<img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
</p>


## 安装 📦

```bash
go get github.com/youfun/cloudflare-d1-go
```

## 快速开始 🚀

### ⭐ 新特性：sqlx 风格 API（推荐使用）

最简洁的查询方式，真正的 sqlx 风格体验：

#### 单行查询 - `Get()`
```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Age   int    `db:"age"`
    Email string `db:"email"`
}

var user User
err := client.Get(
    &user,
    "SELECT * FROM users WHERE name = ?",
    "Alice",  // 直接传递参数，无需 []string
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("用户: %s(年龄: %d)\n", user.Name, user.Age)
```

#### 多行查询 - `Select()`
```go
var users []User
err := client.Select(
    &users,
    "SELECT * FROM users WHERE age > ? ORDER BY age ASC",
    25,  // 直接传递 int 参数
)
if err != nil {
    log.Fatal(err)
}
for _, u := range users {
    fmt.Printf("%s(年龄: %d)\n", u.Name, u.Age)
}
```

#### 执行更新/插入 - `Exec()`
```go
// 执行 UPDATE 并获取受影响的行数
rowsAffected, err := client.Exec(
    "UPDATE users SET age = ? WHERE id = ?",
    30,   // 直接传递 int
    123,  // 直接传递 int
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("更新了 %d 行\n", rowsAffected)
```

#### 多种参数类型（自动转换）
```go
// 字符串参数
err := client.Get(&user, "SELECT * FROM users WHERE name = ?", "Alice")

// 整数参数
err := client.Select(&users, "SELECT * FROM users WHERE age > ?", 25)

// 布尔参数
err := client.Select(&results, "SELECT * FROM users WHERE active = ?", true)

// 时间参数（自动格式化）
startTime := time.Now().Add(-24 * time.Hour)
err := client.Select(&users, "SELECT * FROM users WHERE created_at > ?", startTime)

// 混合参数
err := client.Select(
    &users,
    "SELECT * FROM users WHERE age > ? AND active = ? AND created_at > ?",
    25,         // int
    true,       // bool
    startTime,  // time.Time
)
```

**优势：**
- ✅ 简洁直观的 API，类似 sqlx
- ✅ 自动类型转换（int、bool、time.Time、string 等）
- ✅ 单行查询 - 无需手动迭代行
- ✅ 自动结构体映射，使用 `db` 标签
- ✅ 严格的错误处理
- ✅ 可变参数 - 直接传递值即可

---

### ConnectionPool 配合 sqlx 风格方法（生产环境推荐）

`ConnectionPool` 提供了带自动缓存的 sqlx 风格方法：

```go
pool := cloudflare_d1_go.NewConnectionPool(accountID, apiToken)
pool.SetCacheAge(1 * time.Hour)

err := pool.Connect("database_name")
if err != nil {
    log.Fatal(err)
}

// 现在使用 sqlx 风格方法
var users []User
err = pool.Select(&users, "SELECT * FROM users WHERE age > ?", 25)

var user User
err = pool.Get(&user, "SELECT * FROM users WHERE id = ?", 123)

rowsAffected, err := pool.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 123)
```

---

### 经典 API（仍然支持）

#### 方法 1: 直接客户端（基础）

#### 初始化客户端 🔑

```go
client := cloudflare_d1_go.NewClient("account_id", "api_token")
```

#### 连接到数据库 📁

```go
client.ConnectDB("database_name")
```

#### 查询数据库 🔍

```go
// 执行 SQL 查询，支持可选参数
// query: SQL 查询语句
// params: 参数值数组，对应查询中的 ? 占位符
client.Query("SELECT * FROM users WHERE age > ?", []string{"18"})
```

#### 带参数的示例：
```go
// 查找特定城市的用户
client.Query("SELECT * FROM users WHERE city = ?", []string{"San Francisco"})

// 查找价格范围内的产品
client.Query("SELECT * FROM products WHERE price >= ? AND price <= ?", []string{"10.00", "50.00"})
```

#### 创建表 📄

```go
client.CreateTable("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
```

#### 删除表 🗑️

```go
client.RemoveTable("users")
```

#### 查询特定数据库（方法 2）🔀

```go
client := cloudflare_d1_go.NewClient("account_id", "api_token")
client.QueryDB(databaseID, "SELECT * FROM users", nil)
```

---

### 方法 2: 连接池（推荐 - 类似 sqlx.DB）

`ConnectionPool` 提供了类似 sqlx 的体验，具有自动缓存和连接持久化功能。**推荐在生产环境中使用**，因为它可以将 API 调用减少 99%。

#### 初始化连接池

```go
pool := cloudflare_d1_go.NewConnectionPool("account_id", "api_token")

// 可选：设置缓存有效期（默认 24 小时）
pool.SetCacheAge(1 * time.Hour)
```

#### 连接到数据库（自动缓存）

```go
// 首次调用：进行 API 请求获取数据库 ID（661ms）
// 后续调用：立即从缓存返回（0ms）
err := pool.Connect("database_name")
if err != nil {
    log.Fatalf("连接失败: %v", err)
}
```

#### 执行查询（就像 sqlx）

```go
// 简单查询
res, err := pool.Query("SELECT * FROM users", nil)

// 带参数的查询
res, err := pool.Query("SELECT * FROM users WHERE age > ? AND age < ?", []string{"25", "35"})

// 插入并获取结果信息
res, err := pool.Query("INSERT INTO users (name, age) VALUES (?, ?)", []string{"Alice", "30"})
if err == nil {
    result, _ := res.ToResult()
    lastID, _ := result.LastInsertId()
    fmt.Printf("插入的 ID: %d\n", lastID)
}

// 更新并获取受影响的行数
res, err := pool.Query("UPDATE users SET age = ? WHERE age > ?", []string{"99", "50"})
if err == nil {
    result, _ := res.ToResult()
    affected, _ := result.RowsAffected()
    fmt.Printf("更新了 %d 行\n", affected)
}
```

#### 处理查询结果（就像 sqlx）

```go
// 定义结构体，使用 db 标签
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Age   int    `db:"age"`
    Email string `db:"email"`
}

// 查询并扫描结果
res, err := pool.Query("SELECT id, name, age, email FROM users ORDER BY age ASC", nil)
if err != nil {
    log.Fatalf("查询失败: %v", err)
}

rows, _ := res.ToRows()
defer rows.Close()

for rows.Next() {
    var user User
    rows.StructScan(&user)
    fmt.Printf("%s（年龄: %d）\n", user.Name, user.Age)
}
```

#### 缓存管理

```go
// 检查数据库是否已缓存
if pool.IsCached("database_name") {
    fmt.Println("使用缓存的连接")
} else {
    fmt.Println("进行新的 API 调用")
}

// 查看缓存信息
info := pool.GetCacheInfo("database_name")
fmt.Printf("数据库 ID: %s，缓存时间: %v\n", info.DatabaseID, info.CachedAt)

// 列出所有缓存的数据库
dbList := pool.ListCachedDatabases()
fmt.Printf("缓存的数据库: %v\n", dbList)

// 清除特定缓存
pool.ClearCache("database_name")

// 清除所有缓存
pool.ClearAllCache()
```

#### 多数据库支持

```go
pool := cloudflare_d1_go.NewConnectionPool(accountID, apiToken)

// 连接到多个数据库
pool.Connect("users_db")
pool.Connect("products_db")

// 查询特定数据库（不切换当前数据库）
res, err := pool.QueryDB("users_db", "SELECT * FROM users", nil)

// 或者切换当前数据库
pool.Connect("products_db")  // 将其设为当前
res, err := pool.Query("SELECT * FROM products", nil)  // 使用 products_db
```

#### 性能对比

```
✅ 首次连接（API 调用）:        661.9ms
✅ 后续连接（缓存）:             0ms
✅ 节省: API 调用减少 99%！
```

**针对不同使用场景的推荐设置：**

```go
// Web 服务（短期连接）
pool.SetCacheAge(1 * time.Hour)
pool.SetAutoReconnect(true)

// 长时间运行的批处理任务
pool.SetCacheAge(24 * time.Hour)

// 开发/测试
pool.SetCacheAge(5 * time.Minute)  // 频繁获取最新数据
```

## 高级功能 🔧

### UPSERT 操作（插入或更新）

D1 支持基于 SQLite 的 UPSERT 操作，类似于 PostgreSQL。这对数据同步和数据去重场景非常有用。

#### 场景 1：用户账户同步

从外部数据源同步用户账户时，需要更新现有用户或插入新用户：

```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
    Age   int    `db:"age"`
}

// Upsert 查询 - 如果邮箱存在则更新，不存在则插入
upsertQuery := `
    INSERT INTO users (id, name, email, age) 
    VALUES (?, ?, ?, ?)
    ON CONFLICT(email) DO UPDATE 
    SET name = excluded.name, age = excluded.age;
`

// 同步用户数据
user := User{ID: 100, Name: "Alice", Email: "alice@example.com", Age: 30}
res, err := client.Query(upsertQuery, []string{
    fmt.Sprintf("%d", user.ID),
    user.Name,
    user.Email,
    fmt.Sprintf("%d", user.Age),
})

if err != nil {
    log.Fatalf("Upsert 失败: %v", err)
}

result, _ := res.ToResult()
rowsAffected, _ := result.RowsAffected()
fmt.Printf("同步用户，受影响行数: %d\n", rowsAffected)
```

**优势：**
- ✅ 不需要先检查用户是否存在
- ✅ 原子操作（无竞态条件）
- ✅ 高效的单次查询同步
- ✅ 自动冲突处理

#### 场景 2：数据去重（跳过重复）

从多个数据源导入数据时，需要跳过重复记录：

```go
// 使用 INSERT OR IGNORE - 如果邮箱不存在则插入，存在则忽略
insertOrIgnoreQuery := "INSERT OR IGNORE INTO users (name, email, age) VALUES (?, ?, ?);"

// 尝试插入重复的记录
emails := []string{"bob@example.com", "charlie@example.com", "bob@example.com"}
names := []string{"Bob", "Charlie", "Bob"}
ages := []string{"25", "35", "25"}

for i := 0; i < len(emails); i++ {
    res, err := client.Query(insertOrIgnoreQuery, []string{names[i], emails[i], ages[i]})
    if err != nil {
        log.Fatalf("插入失败: %v", err)
    }
    
    result, _ := res.ToResult()
    rowsAffected, _ := result.RowsAffected()
    
    if rowsAffected > 0 {
        fmt.Printf("✓ 插入用户 %s\n", names[i])
    } else {
        fmt.Printf("⊘ 跳过重复用户 %s\n", names[i])
    }
}
```

**优势：**
- ✅ 自动重复检测
- ✅ 重复插入时不会报错
- ✅ 清晰的批量导入流程
- ✅ 清楚显示插入和跳过的记录

### UPSERT 语法对比

D1 支持多种 UPSERT 方式。**三种方法均已穎试验证：**

```sql
-- 方法 1：INSERT OR IGNORE（跳过重复）✓ 已测试
INSERT OR IGNORE INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?);

-- 方法 2：INSERT ... ON CONFLICT ... DO UPDATE（选择性更新）✓ 已测试
INSERT INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?)
ON CONFLICT(email) DO UPDATE 
SET name = excluded.name, age = excluded.age;

-- 方法 3：INSERT OR REPLACE（替换整行）✓ 已测试
INSERT OR REPLACE INTO users (id, name, email, age) 
VALUES (?, ?, ?, ?);
```

> **测试结果：** 三种 UPSERT 方法均已成功验证：
> - 方法 1：正确跳过重复插入（返回 0 行受影响）
> - 方法 2：正确根据冲突列更新现有记录（返回 1 行受影响）
> - 方法 3：正确查找和替换符合主键的整行（返回 1 行受影响）

## 方法参考 📚

### 数据库管理
- `NewClient(accountID, apiToken string) *Client` - 创建新的 D1 客户端
- `ListDB() (*APIResponse, error)` - 列出账户中的所有数据库
- `CreateDB(name string) (*APIResponse, error)` - 创建新数据库
- `DeleteDB(databaseID string) (*APIResponse, error)` - 删除数据库
- `ConnectDB(name string) error` - 按名称连接到数据库以供后续操作使用

### 表操作
- `CreateTable(createQuery string) (*APIResponse, error)` - 在已连接的数据库中创建表
- `RemoveTable(tableName string) (*APIResponse, error)` - 从已连接的数据库中删除表
- `CreateTableWithID(databaseID, createQuery string) (*APIResponse, error)` - 在特定数据库中创建表
- `RemoveTableWithID(databaseID, tableName string) (*APIResponse, error)` - 从特定数据库中删除表

### 查询执行
- `Query(query string, params []string) (*APIResponse, error)` - 在已连接的数据库上执行查询
  - 支持 SELECT、INSERT、UPDATE、DELETE 等所有 SQL 操作
  - 参数通过数组传递，对应 SQL 中的 `?` 占位符
  - 示例：`client.Query("INSERT INTO users (name, age) VALUES (?, ?)", []string{"Alice", "30"})`
  - 示例：`client.Query("SELECT * FROM users WHERE age > ? AND age < ?", []string{"20", "40"})`
- `QueryDB(databaseID string, query string, params []string) (*APIResponse, error)` - 在特定数据库上执行查询
  - 功能同上，但用于未连接的特定数据库

#### sqlx 风格便利方法（推荐使用）
- `QueryOne(query string, params []string, dest interface{}) error` - 查询单一行并扩描到结构体
  - `dest` 必须是指向结构体的指针，例如 `&user`
  - 如果没有找到行，则返回错误
  - 示例：`client.QueryOne("SELECT * FROM users WHERE id = ?", []string{"1"}, &user)`
- `QueryAll(query string, params []string, dest interface{}) error` - 查询多个行并扩描到结构体切片
  - `dest` 必须是指向切片的指针，例如 `&[]User{}`
  - 如果没有找到行，则返回空切片
  - 示例：`client.QueryAll("SELECT * FROM users", nil, &users)`

#### 常见 SQL 操作示例
**单条插入：**
```go
res, err := client.Query(
    "INSERT INTO users (name, age, email) VALUES (?, ?, ?)",
    []string{"Alice", "30", "alice@example.com"},
)
```

**单条删除：**
```go
res, err := client.Query(
    "DELETE FROM users WHERE id = ?",
    []string{"1"},
)
```

**多条件查询（AND）：**
```go
res, err := client.Query(
    "SELECT * FROM users WHERE age > ? AND age < ? ORDER BY age ASC",
    []string{"25", "35"},
)
```



**多条件更新：**
```go
res, err := client.Query(
    "UPDATE users SET age = ? WHERE age > ? AND name != ?",
    []string{"31", "30", "Alice"},
)
result, _ := res.ToResult()
rowsAffected, _ := result.RowsAffected()
```

**多条件删除：**
```go
res, err := client.Query(
    "DELETE FROM users WHERE age < ? AND status = ?",
    []string{"18", "inactive"},
)
```

### 响应方法
- `ToRows() (*Rows, error)` - 将 SELECT 查询响应转换为 Rows 用于迭代
- `ToResult() (*Result, error)` - 将 INSERT/UPDATE/DELETE 响应转换为 Result 用于获取元数据
- `Get(dest interface{}) error` - 扩描第一行到结构体（sqlx 风格）
  - `dest` 必须是指向结构体的指针
  - 如果没有找到行，则返回错误
- `StructScanAll(dest interface{}) error` - 扩描所有行到结构体切片（sqlx 风格）
  - `dest` 必须是指向切片的指针
  - 如果没有找到行，则返回空切片

### Rows 方法（用于 SELECT 查询）
- `Next() bool` - 料备下一行结果供读取
- `Scan(dest ...interface{}) error` - 将当前行的列复制到目标变量
- `StructScan(dest interface{}) error` - 使用 `db` 标签将当前行扩描到结构体
- `StructScanAll(dest interface{}) error` - 扩描所有一下行到结构体切片（sqlx 风格）
  - `dest` 必须是指向切片的指针
  - 当你已经有了 Rows 对象时很有用
- `Columns() ([]string, error)` - 返回列名
- `Close() error` - 关闭Rows

### Result 方法（用于 INSERT/UPDATE/DELETE）
- `LastInsertId() (int64, error)` - 返回最后插入的行 ID
- `RowsAffected() (int64, error)` - 返回受影响的行数

## 配置 🔧

设置环境变量或使用 `.env` 文件：

```bash
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_API_TOKEN=your_api_token
```

详细说明请参考 `example/.env.example`。

## 示例 📖

查看 `example/main.go` 文件了解完整示例，包括：
- ✅ **sqlx 风格 API** - QueryOne() 和 QueryAll() 示例
- ✅ 批量插入操作
- ✅ 多条件 WHERE 查询（AND、OR）
- ✅ 单行查询与 Get()
- ✅ 多行查询与 StructScanAll()
- ✅ JOIN 查询（LEFT JOIN、INNER JOIN）
- ✅ UPSERT 操作
- ✅ 数据同步场景
- ✅ StructScan 用法

运行示例：
```bash
cd example
cp .env.example .env
# 编辑 .env 并填入你的 Cloudflare 凭证
go run main.go
```

## 迁移管理 🗄️

迁移包为管理数据库架构变更提供了强大的方式。

### 功能特性

- **版本跟踪**: 自动在 `d1_migrations` 表中跟踪已应用的迁移
- **多源支持**: 支持多种迁移源：
  - `FileMigrationSource`: 从本地目录加载（例如 `migrations/`）
  - `EmbedFileSystemMigrationSource`: 使用 `embed.FS` 实现单二进制部署
  - `MemoryMigrationSource`: 内存中的迁移列表
- **SQL 格式**: 兼容 sql-migrate 格式（`-- +migrate Up`、`-- +migrate Down`）
- **D1 集成**: 直接与 cloudflare-d1-go 客户端协作

### 创建迁移文件

在目录中创建 SQL 文件（例如 `migrations/1_init.sql`）：

```sql
-- +migrate Up
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER);
CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT);

-- +migrate Down
DROP TABLE posts;
DROP TABLE users;
```

### 在代码中应用迁移

```go
import (
    "github.com/youfun/cloudflare-d1-go/migrations"
)

func main() {
    // 初始化客户端并连接到数据库
    client := cloudflare_d1_go.NewClient(accountID, apiToken)
    err := client.ConnectDB("database_name")
    if err != nil {
        log.Fatal(err)
    }

    // 定义迁移源
    source := &migrations.FileMigrationSource{
        Dir: "migrations",
    }

    // 应用迁移
    n, err := migrations.Exec(client, source, migrations.Up)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("已应用 %d 个迁移!\n", n)
}
```

### 使用嵌入式迁移

用于单二进制部署：

```go
import (
    "embed"
    "github.com/youfun/cloudflare-d1-go/migrations"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func main() {
    // ... 客户端设置 ...
    
    source := &migrations.EmbedFileSystemMigrationSource{
        FS: migrationFS,
        Dir: "migrations",
    }
    
    n, err := migrations.Exec(client, source, migrations.Up)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 回滚迁移

```go
// 回滚最后一个迁移
n, err := migrations.Exec(client, source, migrations.Down)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("已回滚 %d 个迁移\n", n)
```

## 测试 🧪

```bash
go test -v
```

## 已知限制 ⚠️

### 事务支持

**D1 REST API 不支持原子事务。** 这是平台级别的限制：

- ✅ **支持的功能：**
  - 单条 SQL 语句（SELECT、INSERT、UPDATE、DELETE）
  - 使用 SQLite 语法的 UPSERT 操作（INSERT OR IGNORE、INSERT OR REPLACE、ON CONFLICT）
  - 复杂查询（JOIN、子查询、聚合函数）

- ❌ **不支持的功能：**
  - 跨多个语句的事务（BEGIN/COMMIT/ROLLBACK）
  - 跨多个查询的原子操作
  - 跨多个请求的顺序一致性保证

**原因：** Cloudflare D1 REST API 将每个查询作为独立请求处理。事务支持仅在 **Workers Binding API**（JavaScript/TypeScript）中可用，它可以在单个请求中批处理多个语句。

**解决方案：**
1. **使用 UPSERT 操作** - 用于原子插入或更新场景：
   ```go
   // 在单个查询内是原子操作
   INSERT INTO users (id, name, email) VALUES (?, ?, ?)
   ON CONFLICT(email) DO UPDATE SET name = excluded.name
   ```

2. **设计幂等操作** - 确保应用逻辑可以安全重试失败的查询

3. **使用 Workers Binding API** - 如果需要事务，使用 Cloudflare Workers 的 D1 绑定（JavaScript/TypeScript）

4. **应用级协调** - 实现乐观锁定或版本列来处理并发更新

## 待办事项 📋

- 更好的错误处理 🛡️
- 更全面的测试覆盖 🧪

## 贡献 🤝

欢迎贡献代码！请随时提交 Pull Request。

## 许可证 📄

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 支持 💪

如有任何问题或疑问，请在 GitHub 上提交 Issue。

## 致谢 🙏

- **HTTP 请求层**: 基于 [ashayas/cloudflare-d1-go](https://github.com/ashayas/cloudflare-d1-go)
- **迁移包**: 基于 [github.com/rubenv/sql-migrate](https://github.com/rubenv/sql-migrate)
