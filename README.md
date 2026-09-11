# Sonar Survey Coverage Planner

```bash
docker compose up -d --build
```

`sonar-survey-coverage-planner` 是供海洋测绘团队使用的离线侧扫声呐规划工作台。它管理投影测区、平行测线、导入航迹和不可覆盖的缺口快照，并为人工补测决策提供可解释几何证据。系统不连接船舶、声呐或自动驾驶设备，不产生控制指令。

## 主要功能

- 测区：创建、校验米制投影边界，查看规划、运行和覆盖摘要。
- 测线：从测区生成平行测线，锁定执行版本，复制形成后续草稿。
- 航迹：导入 GeoJSON，检查采样点、长度、航速与导航质量，按状态机处理。
- 覆盖：以固定网格估算覆盖、重复覆盖和漏测，冻结输入哈希并生成补测线建议。
- 审计：记录四类实体写操作的前后快照、操作者、角色、request ID 和算法元数据。

## 角色与账号

所有演示账号初始密码均为 `Sonar2026!`。

| 账号 | 角色 | 权限 |
| --- | --- | --- |
| `admin` | `admin` | 测区、规划、航迹和覆盖计算管理 |
| `planner` | `survey_planner` | 测区与测线规划 |
| `processor` | `data_processor` | 航迹导入、处理和覆盖计算 |
| `reviewer` | `reviewer` | 覆盖缺口人工复核与审计读取 |
| `auditor` | `auditor` | 全局只读与审计读取 |

审计员在数据库角色、JWT claims、Gin RBAC、React 路由和按钮层均为只读。只有 `reviewer` 可以推进缺口复核状态。

## 页面

| 路径 | 主要实体 | 交互 |
| --- | --- | --- |
| `/areas` | SurveyArea、TransectPlan | 创建投影测区、查看覆盖摘要和边界 |
| `/plans` | TransectPlan、SurveyArea | 生成平行测线、锁定或复制版本 |
| `/runs` | SonarRun、TransectPlan | 导入航迹、读取质量证据、推进处理状态 |
| `/coverage` | CoverageGap、SonarRun、SurveyArea | 计算覆盖、查看缺口与补测线、人工复核 |
| `/audit` | 四实体审计投影 | 按 request ID、实体和操作者筛选 |

所有页面通过 `/api/v1` 读取真实数据。二维测绘画布使用本地 Canvas，不依赖在线地图或第三方瓦片服务。

## 架构

```text
React 18 + Material UI + Zustand
              |
          Nginx /api
              |
Gin handlers -> services -> repositories -> GORM
                                      |       |
                              PostgreSQL 16  SQLite smoke
```

后端四实体分别拆分为 `model`、`dto`、`repository`、`service` 和 `handler` 文件；前端分别拆分为 `type`、`api`、`store` 和页面消费。几何解析使用 `paulmach/orb/geojson`，正式数据库启用 PostGIS 扩展。

## 目录

```text
backend/cmd/server        服务入口与优雅停机
backend/internal/config  配置、数据库迁移与种子
backend/internal/geometry 结构化 GeoJSON 与覆盖算法
backend/internal/{model,dto,repository,service,handler}
backend/internal/middleware JWT、RBAC、request ID、恢复、审计、错误
backend/internal/router  API 路由与限流
backend/pkg/api          统一成功和错误响应
frontend/src/{types,api,stores,pages}
frontend/src/components/common 共享几何与状态组件
database/init.sql        PostGIS 扩展初始化
```

## API

所有业务 API 使用 `/api/v1`，健康端点为 `/healthz`。

| Method | Path | 说明 |
| --- | --- | --- |
| POST | `/auth/login` | 登录并取得 JWT |
| GET | `/auth/me` | 当前身份 |
| GET/POST | `/areas` | 测区列表与创建 |
| GET/PUT | `/areas/:id` | 测区详情与版本更新 |
| GET/POST | `/plans` | 规划列表与手工创建 |
| POST | `/plans/generate` | 从测区生成平行测线 |
| PUT | `/plans/:id` | 更新草稿规划 |
| POST | `/plans/:id/transition` | 锁定规划 |
| POST | `/plans/:id/copy` | 复制新版本 |
| GET | `/runs`、`/runs/:id` | 运行列表与详情 |
| POST | `/runs/import` | 导入航迹，按 checksum 幂等 |
| GET | `/runs/:id/quality` | 航迹质量证据 |
| POST | `/runs/:id/transition` | 推进运行状态机 |
| GET | `/coverage-gaps`、`/coverage-gaps/:id` | 缺口快照列表与详情 |
| POST | `/coverage-gaps/detect` | 覆盖计算，要求 `Idempotency-Key` |
| POST | `/coverage-gaps/:id/transition` | reviewer 人工复核 |
| GET | `/audits` | 审计筛选 |

错误响应统一包含业务 `code`、`message`、可选 `details` 和 `request_id`。无效 GeoJSON/坐标系返回 422，非法状态或版本冲突返回 409，认证与权限分别返回 401/403。

## 共享枚举位置

`RunState = imported | quality_checked | processing | processed | rejected | superseded`

- 后端：`internal/constants/run_state.go`；`model/sonar_run.go`；`dto/sonar_run.go`；`service/sonar_run.go` 状态机；`handler/sonar_run.go`；`router/router.go`；`constants/state_test.go`。
- 前端：`types/enums/run-state.ts`；`types/sonar-run.ts`；`stores/sonar-run-store.ts`；`components/common/RunStateBadge.tsx`；`pages/RunsPage.tsx`、`pages/CoveragePage.tsx`；`utils/state.test.ts`。

`GapSeverity = minor | major | critical`

- 后端：`internal/constants/gap_severity.go`；`model/coverage_gap.go`；`dto/coverage_gap.go`；`service/coverage_gap.go`；`handler/coverage_gap.go`；`constants/state_test.go`。
- 前端：`types/enums/gap-severity.ts`；`types/coverage-gap.ts`；`stores/coverage-gap-store.ts`；`pages/CoveragePage.tsx`；`utils/state.test.ts`。

## 坐标与算法边界

- 面积、距离和扫幅计算只接受以米为单位的项目投影坐标，明确拒绝 EPSG:4326/WGS84。
- 边界和航迹使用结构化 GeoJSON 解析；固定夹具测试覆盖 Polygon、MultiLineString、面积和稳定哈希。
- 当前算法以目标分辨率构造有限网格，用点到线段距离近似扫幅覆盖，计算覆盖率、重叠率和漏测率。
- 小于分辨率阈值的碎片会在解释中计数；补测线沿缺口包围盒主方向生成。
- 结果是离线规划近似，不替代水深、海况、导航误差和持证测绘人员判断，也不能下发船舶控制。

## 环境变量与端口

复制 `.env.example` 为本地 `.env` 后调整密码。仓库中的 `.env` 不会被 Git 提交。

| 变量 | Compose 默认值 | 用途 |
| --- | --- | --- |
| `COMPOSE_PROJECT_NAME` | `sonar-survey-coverage-planner` | 固定英文资源前缀 |
| `FRONTEND_PORT` | `18532` | Web 入口 |
| `BACKEND_PORT` | `19532` | 后端直连 |
| `DB_PORT` | `57532` | PostgreSQL 宿主端口 |
| `DB_DRIVER` | `postgres` | 数据库驱动 |
| `DB_NAME` / `DB_USER` / `DB_PASSWORD` | 见 `.env.example` | 数据库身份 |
| `JWT_SECRET` | 必填 | JWT HMAC 密钥，至少 24 字符 |

Redis、MinIO 均未使用，也没有额外宿主机端口。

## 本地开发

```bash
go work sync
DB_DRIVER=sqlite DB_DSN='file:local.db?_foreign_keys=on' JWT_SECRET='local-development-secret-change-me' go run ./backend/cmd/server
npm --prefix frontend ci --registry=https://registry.npmjs.org --replace-registry-host=always
npm --prefix frontend run dev
```

Vite 开发服务器把 `/api` 代理到 `127.0.0.1:20532`。也可以按 `runtime_smoke.json` 直接启动后端 smoke 服务。

## Runtime smoke

```bash

```

manifest 在 `backend/` 使用 SQLite 内存数据库启动 `go run ./cmd/server`，并等待 `http://127.0.0.1:20532/healthz` 返回 200。

## 构建与测试

```bash
go work sync
go build ./backend/...
go vet ./backend/...
go test ./backend/...
go test -race ./backend/...
npm --prefix frontend ci --registry=https://registry.npmjs.org --replace-registry-host=always
npm --prefix frontend test
npm --prefix frontend run typecheck
npm --prefix frontend run build
npm --prefix frontend audit --registry=https://registry.npmjs.org
docker compose config --quiet
```

## Docker 部署与停止

```bash
docker compose up -d --build
docker compose ps
docker compose down -v --remove-orphans
```

数据库、后端和前端均配置 healthcheck。后端仅在 PostgreSQL healthy 后启动，前端仅在后端 healthy 后启动。Nginx 原样代理 `/api/v1`，并用 `/api/healthz` 暴露后端健康状态。

## 常见问题

- `COORDINATE_SYSTEM_INVALID`：把经纬度转换为项目约定的米制投影坐标后重新创建测区。
- `VERSION_CONFLICT`：数据已被其他人员更新，刷新列表后按新版本重试。
- `RUN_TRANSITION_INVALID`：必须依次完成质量检查、处理和已处理状态。
- `RUN_NOT_PROCESSED`：覆盖计算只能选择已处理且属于同一测区的运行。
- npm 默认镜像无法下载或审计：显式使用 `--registry=https://registry.npmjs.org --replace-registry-host=always`。

## License

MIT License，详见 `LICENSE`。
