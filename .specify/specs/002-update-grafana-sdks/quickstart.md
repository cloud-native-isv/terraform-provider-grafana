# Quickstart: Update Grafana SDKs

## 目标
将 Grafana App SDK 升级到 `v0.50.1`、Grafana Plugin SDK Go 升级到 `v0.287.0`，并验证标准构建与关键使用场景兼容。

## 步骤
1. 在 `go.mod` 中更新依赖版本。
2. 运行 `go mod tidy` 以同步 `go.sum`。
3. 运行 `go build` 或 `make build` 确认构建通过。
4. 运行关键使用场景验证（至少执行既有核心资源/数据源的最小路径）。
5. 如需更新生成文档，运行 `go generate ./...`（注意 `docs/` 禁止手工编辑）。

## 验证清单
- App SDK 版本为 `v0.50.1`
- Plugin SDK Go 版本为 `v0.287.0`
- 标准构建流程成功
