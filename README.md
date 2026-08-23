# 商场消防设施巡检与隐患整改平台

本项目连接设备台账、路线排班、扫码巡检、隐患整改、复验恢复、备件和合规审计。Go 服务按领域拆分，Vue 前端提供中文运营界面，所有外部能力均有本地适配器。

## 架构

- `internal/domain`：组织权限、设备事件、模板快照、路线排班、巡检任务、隐患状态机、备件与报告。
- `internal/application`：事务用例及靠近调用方的仓储、时钟、文件、通知和 outbox 接口。
- `internal/repository`：内存仓储和 pgx/PostgreSQL 仓储。
- `internal/transport/http`：`/api/v1` 路由、校验与稳定错误 envelope。
- `internal/platform`：本地时钟、ID、文件、幂等、通知和可取消重试 outbox。
- `migrations`、`api/openapi`、`web` 和 `deploy`：数据库、接口、前端与部署资产。

巡检任务在生成时复制模板版本，后续模板发布不改写历史任务。扫码提交验证任务、设备、巡检人、二维码和有效时间窗，并以任务与设备组合保证离线补传幂等。严重隐患必须经过复验；接受复验时隐患关闭、设备恢复事件、资产状态和审计在一个事务中提交。设备当前状态由最后一条有效事件推导。

## 本地运行

需要 Go 1.24+。默认以内存仓储运行，完全离线可演示：

```text
go run ./cmd/server
```

健康检查位于 `/healthz` 和 `/readyz`。业务请求使用 `Authorization`、`X-Actor-ID`、`X-Mall-ID`、`X-Team-ID` 和 `X-Roles` 演示服务端身份范围。

## 容器和数据库

`docker compose up --build` 启动应用与 PostgreSQL。镜像采用多阶段构建和非 root 用户，Go 交叉编译支持 amd64/arm64。设置 `DATABASE_URL` 后执行 `scripts/migrate.ps1` 应用版本化迁移和幂等种子。

## 验证

```text
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

前端执行 `npm test`、`npm run typecheck` 和 `npm run build`。演示实现不连接真实短信或推送，通知与附件保存在本地适配器；访问令牌入口是可替换演示中间件，生产部署应接入受控密钥和持久化刷新令牌。
