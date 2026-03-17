# Quickstart: grafana_datasource_query

## Prerequisites

- 已配置 `grafana` provider（`url` + `auth`）
- 目标 Grafana 数据源已存在且凭据具备查询权限
- 当前分支包含 `grafana_datasource_query` 资源实现

## 1) 最小示例（managed 模式）

```hcl
provider "grafana" {
  url  = var.grafana_url
  auth = var.grafana_auth
}

resource "grafana_datasource_query" "managed_cpu" {
  mode          = "managed"
  datasource_id = 1

  managed_query {
    instant = false
    queries = [
      {
        ref_id = "A"
        expr   = "sum(rate(node_cpu_seconds_total[5m]))"
      }
    ]
  }

  # 可省略，省略时默认 now-24h ~ now（UTC 分钟对齐）
  time_range {
    from = "now-24h"
    to   = "now"
  }
}

output "managed_summary" {
  value = grafana_datasource_query.managed_cpu.result_summary
}
```

## 2) 最小示例（proxy 模式）

```hcl
resource "grafana_datasource_query" "proxy_prom" {
  mode          = "proxy"
  datasource_id = 1

  proxy_query {
    path   = "/api/v1/query"
    method = "GET"
    params = {
      query = "up"
      time  = "now"
    }
  }
}
```

## 3) 执行与验证

```bash
terraform init
terraform apply
terraform refresh
terraform plan
```

验证点：
- `result_summary` 存在且字段结构稳定
- 重复执行且上游结果等价时，不出现无意义状态变更
- 失败时返回可分类错误（validation/auth/upstream/network/throttling）

## 4) 失败诊断建议

- `validation`：检查 `mode` 与对应块（`managed_query`/`proxy_query`）是否匹配
- `authorization`：检查 provider `auth` 与数据源访问权限
- `throttling`/`upstream`：确认上游健康与限流状态；资源会指数退避最多 3 次
- `network`：检查网络连通、TLS 与代理设置

## 5) 质量门禁（实现后）

```bash
go test ./internal/resources/grafana/...
go test ./pkg/provider/...
# 根据仓库规范运行 acceptance（按环境选择对应 target）
make testacc
```