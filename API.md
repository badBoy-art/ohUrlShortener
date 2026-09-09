# ohUrlShortener HTTP API

### `/api` 接口权限说明

所有 `/api/*` 接口需要通过 `Bearer Token` 方式验证权限，亦即：每个请求 Header 须携带 

```shell
 Authorization: Bearer {sha256_of_password}
```

`sha256_of_password` 的加密规则，与 `storage/users_storage.go` 中的 `PasswordBase58Hash()` 保持同步

### 1. 新增短链接 `POST /api/url`

接受参数：
1. `dest_url` 主目标链接，选填（与 `destinations` 至少提供一个；两者都提供时以 `dest_url` 为主目标）
2. `memo` 备注信息，选填
3. `open_type` 打开方式（0-8），选填
4. `destinations` 多目标地址列表（JSON 数组），选填，最多 20 个；每个元素包含 `label`（目标标识，最长64字符，唯一）与 `dest_url`（目标链接，最长2048字符）

请求示例：

```shell
curl --request POST \
  --url http://localhost:9092/api/url \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data dest_url=http://localhost:9092/admin/dashboard \
  --data memo=dashboard
```

多目标短链接示例（PC 与移动端各一个目标地址）：

```shell
curl --request POST \
  --url http://localhost:9092/api/url \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'dest_url=https://www.example.com' \
  --data-urlencode 'destinations=[{"label":"pc","dest_url":"https://www.example.com"},{"label":"mobile","dest_url":"https://m.example.com"}]' \
  --data memo=dashboard
```

也可以不传 `dest_url`，此时取 `destinations` 中第一个目标地址作为主目标（用于生成短码与默认跳转）：

```shell
curl --request POST \
  --url http://localhost:9092/api/url \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --data-urlencode 'destinations=[{"label":"pc","dest_url":"https://www.example.com"},{"label":"mobile","dest_url":"https://m.example.com"}]'
```

返回结果：

```shell
{
	"code": 200,
	"status": true,
	"message": "success",
	"result": {
		"short_url": "http://localhost:9091/BUUtpbGp"
	},
	"date": "2022-04-10T21:31:29.36559+08:00"
}
```

#### 多目标短链接访问方式

访问短链接时通过请求头 `X-Dest-Label` 指定目标地址标识（即创建时的 `label`）；未携带请求头或标识未命中时，回退到主目标地址 `dest_url`（旧数据同样如此，完全兼容）：

```shell
# 302 跳转到 mobile 对应的目标地址
curl -i --header 'X-Dest-Label: mobile' http://localhost:9091/BUUtpbGp

# 不带请求头，跳转到主目标地址
curl -i http://localhost:9091/BUUtpbGp
```

### 2. 禁用/启用 短链接 `PUT /api/url/:url/change_state`

接受参数：
1. `url` path 参数，指定短链接，必填
2. `enable` 禁用时，传入 false；启用时，传入 true 

请求示例：

```shell
curl --request PUT \
  --url http://localhost:9092/api/url/33R5QUtD/change_state \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data enable=false
```

返回结果：

```shell
{
	"code": 200,
	"status": true,
	"message": "success",
	"result": true,
	"date": "2022-04-10T21:31:25.7744402+08:00"
}
```

### 3. 查询短链接统计数据 `GET /api/url/:url`

接受参数：
1. `url` path 参数，指定短链接，必填

请求示例：

```shell
curl --request GET \
  --url http://localhost:9092/api/url/33R5QUtD \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded'
```

返回结果：

```shell
{
	"code": 200,
	"status": true,
	"message": "success",
	"result": {
		"short_url": "33R5QUtD",
		"today_count": 3,
		"yesterday_count": 0,
		"last_7_days_count": 0,
		"monthly_count": 3,
		"total_count": 3,
		"d_today_count": 1,
		"d_yesterday_count": 0,
		"d_last_7_days_count": 0,
		"d_monthly_count": 1,
		"d_total_count": 1
	},
	"date": "2022-04-10T21:31:22.059596+08:00"
}
```

### 4. 新建管理员 `POST /api/account`

接受参数：
1. `account` 管理员帐号，必填
2. `password` 管理员密码，必填，最小长度8

请求示例：

```shell
curl --request POST \
  --url http://localhost:9092/api/account \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data account=hello1 \
  --data password=12345678
```

返回结果：

```shell
{
	"code": 200,
	"status": true,
	"message": "success",
	"result": null,
	"date": "2022-04-10T21:31:39.7353132+08:00"
}
```

### 5. 修改管理员密码 `PUT /api/account/:account/update`

接受参数：
1. `account` path 参数，管理员帐号，必填
1. `password` 管理员密码，必填，最小长度8

请求示例：

```shell
curl --request PUT \
  --url http://localhost:9092/api/account/hello/update \
  --header 'Authorization: Bearer EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data password=world123
```

返回结果：

```shell
{
	"code": 200,
	"status": true,
	"message": "success",
	"result": null,
	"date": "2022-04-10T21:31:32.5880538+08:00"
}
```

### 6. 删除短链接 `DELETE /api/url/:url`

接受参数：
1. `url` path 参数，要删除的短链接地址

（此处省略示例）