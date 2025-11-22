# 环境配置指南

## 快速开始

### 1. 复制配置文件模板

```bash
cp .env.example .env
```

### 2. 获取 Cloudflare 凭证

#### Account ID
1. 访问 [Cloudflare Dashboard](https://dash.cloudflare.com/?to=/:account/overview)
2. 在右侧"API"部分找到你的 **Account ID**
3. 复制并粘贴到 `.env` 文件中

#### API Token
1. 访问 [API Tokens 页面](https://dash.cloudflare.com/profile/api-tokens)
2. 点击 "Create Token"
3. 选择 "Edit D1" 或创建自定义权限
4. 所需权限：`D1:Edit`
5. 复制生成的 Token 到 `.env` 文件

#### Database Name
1. 在 Cloudflare Dashboard 中找到你的 D1 数据库
2. 复制数据库名称到 `.env` 文件

### 3. 配置 .env 文件

编辑 `.env` 文件（在 `example/` 目录中）：

```
CLOUDFLARE_ACCOUNT_ID=your_actual_account_id
CLOUDFLARE_API_TOKEN=your_actual_api_token
CLOUDFLARE_DB_NAME=your_actual_database_name
```

### 4. 运行示例

```bash
go run main.go
```

## 优先级说明

程序会按以下优先级加载配置：

1. **环境变量** （最高优先级）
   ```bash
   export CLOUDFLARE_ACCOUNT_ID=xxx
   go run main.go
   ```

2. **.env 文件** （次优先级）
   - 自动从当前目录或上层目录读取 `.env` 文件
   - 不会覆盖已设置的环境变量

## 安全建议

⚠️ **重要**：
- ✅ `.env.example` 可以提交到 Git
- ❌ `.env` 文件包含敏感信息，**不要提交**到版本控制
- ❌ 不要在代码中硬编码凭证
- ✅ 使用 `.gitignore` 排除 `.env` 文件

## 验证配置

如果配置有问题，程序会输出类似信息：

```
Please set CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_API_TOKEN, and CLOUDFLARE_DB_NAME environment variables
```

检查：
1. `.env` 文件是否存在且格式正确
2. 环境变量是否正确设置
3. Cloudflare 凭证是否有效

## 迁移文件配置

项目包含自动化迁移系统，用于版本控制数据库架构。迁移文件位于 `../migrations/` 目录。

### 迁移工作原理

1. **自动跟踪**: 首次运行时，系统创建 `d1_migrations` 表来跟踪已应用的迁移
2. **增量应用**: 每次运行时，仅应用新的迁移文件
3. **版本控制**: 迁移文件按顺序编号（例如 `1_init.sql`、`2_add_columns.sql`）

### 迁移文件格式

迁移文件使用标准 SQL 格式，包含 Up 和 Down 部分：

```sql
-- +migrate Up
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);

-- +migrate Down
DROP TABLE users;
```

### 在应用中应用迁移

参考 `main.go` 中的迁移示例：

```go
source := &migrations.FileMigrationSource{
    Dir: "../migrations",  // 迁移文件目录
}

n, err := migrations.Exec(client, source, migrations.Up)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("已应用 %d 个迁移\n", n)
```

### 常见问题

**Q: 为什么第二次运行时没有创建表？**  
A: 这是正常的！迁移系统会跟踪已应用的迁移。第一次运行时应用了迁移，第二次运行时检测到迁移已应用，会跳过创建步骤。

**Q: 如何修改已应用的迁移？**  
A: 不要修改已应用的迁移文件。而是创建新的迁移文件来进行增量更改：
- `1_init.sql` - 原始表结构
- `2_add_columns.sql` - 添加新列
- `3_create_indexes.sql` - 创建索引

**Q: 如何回滚迁移？**  
A: 使用 `migrations.Down` 来回滚最后应用的迁移：

```go
n, err := migrations.Exec(client, source, migrations.Down)
```
