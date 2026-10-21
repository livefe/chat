# Nearby Chat

基于地理位置的附近社交平台，包含 H5 端、Android 端、管理后台和服务端。

用户可以在附近的聊天室和周边广场里交流，发布活动、通知、闲置、求助等本地动态，并通过私信与客服继续沟通；运营人员在管理后台完成内容审核、用户治理和数据运营。

## 目录

- [功能特性](#功能特性)
- [架构](#架构)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
  - [环境要求](#环境要求)
  - [获取代码](#获取代码)
  - [启动服务端和前端](#启动服务端和前端)
  - [运行 Android 端](#运行-android-端)
- [开发规范](#开发规范)
- [许可证](#许可证)

## 功能特性

**H5 端 / Android 端**

- 附近聊天：按位置进入聊天室，或在周边广场与附近的人公开交流
- 动态：发布活动、通知、闲置、求助，支持图片、地点、活动报名与签到、动态内讨论
- 私信：一对一聊天，支持图片、语音、视频、文件、位置、回复、撤回、转发
- 会话管理：置顶、删除、未读计数，离线时通过小米推送提醒
- 用户关系：个人主页、拉黑、举报、意见反馈
- 在线客服：用户与客服坐席实时对话
- 分享链接：H5 分享链接可唤起 Android App（App Links / 自定义 scheme）

**管理后台**

- 内容审核：动态、聊天室、广告、举报的审核与处置
- 用户治理：用户、消息、会话管理，敏感词库
- 组织与权限：部门、员工、角色、菜单权限（RBAC），关键操作留有审计日志
- 运营：广告投放与统计、App 版本发布与更新策略、行政区划
- 员工即时通讯与客服工作台

**服务端**

- REST API 和 WebSocket 实时通道，可水平扩展为多实例，实例之间通过 Redis Pub/Sub 互通
- 用户与员工两套 JWT 鉴权，支持令牌吊销和刷新
- 消息发送滑动窗口限流、敏感词过滤、请求日志脱敏
- 独立的 worker 进程负责定时任务，使用分布式锁避免重复执行
- 可插拔的短信（阿里云 / mock）和对象存储（腾讯云 COS / 本地磁盘）

## 架构

```
  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
  │   chat-h5    │   │ chat-android │   │  chat-admin  │
  │   (Vue 3)    │   │   (Kotlin)   │   │   (Vue 3)    │
  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
         └──────────────────┼──────────────────┘
                            │ HTTP / WebSocket
  ┌─ chat-server ───────────┼──────────────────────────┐
  │               ┌─────────▼─────────┐  ┌───────────┐ │
  │               │      server       │  │  worker   │ │
  │               │     API / WS      │  │ cron jobs │ │
  │               └─────────┬─────────┘  └─────┬─────┘ │
  └─────────────────────────┼──────────────────┼───────┘
                            │                  │
         ┌──────────────────┼──────────────────┤
  ┌──────▼───────┐   ┌──────▼───────┐   ┌──────▼───────┐
  │    MySQL     │   │    Redis     │   │ SMS, Storage │
  │   database   │   │ cache/pubsub │   │    & Push    │
  └──────────────┘   └──────────────┘   └──────────────┘
```

## 技术栈

| 目录 | 端 | 技术 |
|------|----|------|
| `chat-h5/` | H5 端 | Vue 3、TypeScript、Vite、Pinia、Vue Router、Vant 4、天地图 |
| `chat-android/` | Android 端 | Kotlin、Jetpack Compose、Hilt、Retrofit、OkHttp、kotlinx.serialization、osmdroid |
| `chat-admin/` | 管理后台 | Vue 3、TypeScript、Vite、Pinia、Vue Router、Element Plus、高德地图 |
| `chat-server/` | 服务端 | Go 1.24、Gin、GORM、MySQL 8、Redis 7、gorilla/websocket、zap、Viper、robfig/cron |
## 快速开始

用 Docker Compose 一次启动服务端、H5 端和管理后台，Android 端在本机构建后连接这套环境。`docker-compose.yml` 仅用于本地开发，不是生产部署方案。

### 环境要求

| 依赖 | 版本 | 用于 |
|------|------|------|
| Docker | 含 Compose V2 | 启动服务端、H5 端和管理后台 |
| JDK | 17+ | 构建 Android 端 |
| Android SDK | API 35 | 构建 Android 端 |
| Go | 1.24+ | 服务端开发检查（可选） |
| Node.js | 24+ | H5 端、管理后台的 lint 和类型检查（可选） |

### 获取代码

```bash
git clone https://github.com/livefe/chat.git
cd chat
```

### 启动服务端和前端

1. 在 `chat-h5/.env.local` 填写 `VITE_TIANDITU_TK`，在 `chat-admin/.env.local` 填写 `VITE_AMAP_KEY` 和 `VITE_AMAP_SECURITY`（地图 Key，不填时地图不可用）。
2. 启动：

   ```bash
   docker compose up -d --build
   ```

3. 首次启动后导入种子数据：

   ```bash
   docker cp chat-server/scripts/init.sql chat-mysql:/tmp/init.sql
   docker compose exec -T mysql mysql -uroot -p123456 --default-character-set=utf8mb4 app -e "SOURCE /tmp/init.sql"
   ```

启动后访问 http://localhost/ 打开 H5 端，http://localhost/admin/ 打开管理后台。管理后台默认账号的手机号是 `17000000000`，验证码固定为 `111111`。

常用命令：

```bash
docker compose ps          # 查看服务状态
docker compose logs -f     # 查看日志
docker compose down        # 停止并移除容器（保留数据）
docker compose down -v     # 同时清空数据卷
```

### 运行 Android 端

```bash
cd chat-android
cp local.properties.example local.properties
```

在 `local.properties` 中填写 `sdk.dir`。服务地址默认指向模拟器里的电脑本机（`http://10.0.2.2/`），真机改成电脑的局域网 IP。

用 Android Studio 打开 `chat-android` 运行，或者用命令行构建并安装到已连接的设备：

```bash
./gradlew :app:installDebug     # Windows 下用 gradlew.bat
```

## 开发规范

提交代码前，请确保对应模块的检查都能通过：

```bash
# chat-server
gofmt -l . && go build ./... && go vet ./...

# chat-h5 / chat-admin
npm run format:check && npm run lint && npm run type-check   # 用 npm run format 自动格式化

# chat-android
./gradlew spotlessCheck   # 用 spotlessApply 自动格式化
```

## 许可证

本项目尚未选择开源许可证。在仓库添加 `LICENSE` 文件之前，默认保留所有权利。
