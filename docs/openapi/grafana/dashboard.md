## 🛠️ Grafana Dashboard HTTP API 开发文档

### 1. 新仪表盘 API (New Dashboard APIs)
> **注意**：这是基于 Kubernetes 风格的 API 结构（v1beta1）。若使用 Grafana Enterprise，部分端点需要特定权限。详情请参阅 [RBAC 权限](#required-permissions-rbac)。

#### 1.1 创建仪表盘 (POST)
**Endpoint**
```http
POST /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards
```
**描述**
创建一个新的仪表盘。

**路径参数**
*   `namespace`：命名空间（参考 [API 概述](/docs/grafana/latest/developers/http_api/apis/)）。

**请求头**
*   `Content-Type: application/json`
*   `Authorization: Bearer <your-token>`

**JSON Body Schema**
| 字段 | 描述 |
| :--- | :--- |
| `metadata.name` | **必填**。仪表盘的 UID。若不想指定，可设置 `metadata.generateName` 作为前缀（不能为字符串）。 |
| `metadata.annotations.grafana.app/folder` | **选填**。仪表盘存放的文件夹 UID。 |
| `spec` | **必填**。仪表盘的 JSON 模型。 |

**响应状态码**
*   `201 Created`：创建成功。
*   `400 Bad Request`：请求错误（JSON 格式无效、字段缺失等）。
*   `401 Unauthorized`：未授权。
*   `403 Forbidden`：访问被拒绝。
*   `409 Conflict`：冲突（同 UID 的仪表盘已存在）。

#### 1.2 更新仪表盘 (PUT)
**Endpoint**
```http
PUT /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards/:uid
```
**描述**
通过 UID 更新现有的仪表盘。

**路径参数**
*   `namespace`：命名空间。
*   `uid`：要更新的仪表盘的唯一标识符（对应响应中的 `metadata.name`）。

**JSON Body Schema**
| 字段 | 描述 |
| :--- | :--- |
| `metadata.name` | **必填**。仪表盘的 UID。 |
| `metadata.annotations.grafana.app/folder` | **选填**。更新仪表盘所在的文件夹 UID。 |
| `metadata.annotations.grafana.app/message` | **选填**。设置版本历史的提交信息（commit message）。 |
| `spec` | **必填**。更新后的仪表盘 JSON 模型。 |

**响应状态码**
*   `200 OK`：更新成功。
*   `409 Conflict`：冲突（仪表盘版本不匹配，需获取最新版本重试）。

#### 1.3 获取仪表盘 (GET)
**Endpoint**
```http
GET /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards/:uid
```
**描述**
通过 UID 获取仪表盘详情。

**查询参数**
*   `namespace`：命名空间。
*   `uid`：仪表盘的 UID。

**附加端点：检索访问信息**
```http
GET /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards/:uid/dto
```
该接口返回额外的 `access` 字段，包含用户对该仪表盘的权限（如是否为管理员、编辑者）或是否为公开仪表盘。

**响应状态码**
*   `200 OK`：成功获取。
*   `404 Not Found`：未找到仪表盘。

#### 1.4 列出仪表盘 (GET)
**Endpoint**
```http
GET /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards
```
**描述**
列出组织中的所有仪表盘，支持分页。

**查询参数**
*   `limit` (可选)：返回仪表盘的最大数量。
*   `continue` (可选)：从上一次响应的 `metadata.continue` 中获取的 token，用于获取下一页。

**响应示例**
```json
{
  "kind": "DashboardList",
  "metadata": {
    "continue": "eyJvIjoxNTIsInYiOjE3NjE3MDQyMjQyMDcxODksInMiOmZhbHNlfQ=="
  },
  "items": [
    {
      "kind": "Dashboard",
      "metadata": { "name": "gpqcmf", ... },
      "spec": { "title": "New dashboard", ... }
    }
  ]
}
```
**分页逻辑**：使用返回的 `continue` token 作为参数循环请求，直到响应的 `metadata` 中不再包含 `continue` 字段，表示已到达最后一页。

#### 1.5 删除仪表盘 (DELETE)
**Endpoint**
```http
DELETE /apis/dashboard.grafana.app/v1beta1/namespaces/:namespace/dashboards/:uid
```
**描述**
通过 UID 删除仪表盘。

**路径参数**
*   `uid`：仪表盘的 UID（对应 `metadata.name`，而非 `metadata.uid`）。

**响应状态码**
*   `200 OK`：删除成功。

---

### 2. 传统仪表盘 API (Legacy Dashboard API)
> **注意**：这是旧版 API，主要用于兼容性。推荐新项目使用上述新 API。

#### 2.1 创建或更新仪表盘 (POST)
**Endpoint**
```http
POST /api/dashboards/db
```
**描述**
全能型端点，根据 `dashboard.id` 和 `overwrite` 字段决定行为。

**JSON Body Schema**
| 字段 | 描述 |
| :--- | :--- |
| `dashboard` | 完整的仪表盘模型。 |
| `dashboard.id` | `null` 表示创建新仪表盘；指定 ID 表示更新现有仪表盘。 |
| `dashboard.uid` | **选填**。创建时可指定 UID，为 `null` 则自动生成。 |
| `folderId` / `folderUid` | 仪表盘保存的文件夹 ID 或 UID（`folderUid` 优先级更高）。若未定义，仪表盘将移至根目录。 |
| `overwrite` | `true` 表示允许覆盖写入（处理版本冲突）。 |
| `message` | 提交信息（版本历史记录）。 |

**响应状态码**
*   `200 OK`：成功（包含 `id`, `uid`, `url`, `status` 等信息）。
*   `412 Precondition Failed`：前置条件失败（常见原因：`version-mismatch` 版本冲突，`name-exists` UID 已存在，`plugin-dashboard` 属于插件仪表盘无法修改）。

#### 2.2 获取仪表盘 (GET)
**Endpoint**
```http
GET /api/dashboards/uid/:uid
```
**描述**
通过 UID 获取仪表盘。

**响应结构**
返回对象包含 `dashboard`（仪表盘 JSON）和 `meta`（元数据，如 `folderId`, `folderUid`, `isStarred` 等）。

#### 2.3 删除仪表盘 (DELETE)
**Endpoint**
```http
DELETE /api/dashboards/uid/:uid
```
**描述**
通过 UID 删除仪表盘。

---

### 3. 核心概念与权限

#### 3.1 唯一标识符 (UID) vs 标识符 (ID)
*   **UID (Unique Identifier)**：
    *   推荐使用。
    *   长度不超过 40 字符。
    *   用于在组织内唯一标识仪表盘，支持一致的访问 URL（即使重命名也不变）。
    *   适用于仪表盘同步和书签。
*   **ID (Identifier)**：
    *   传统数据库自增 ID。
    *   **已弃用**。

#### 3.2 权限控制 (RBAC)
以下是新 API 结构中涉及的主要权限：

| 操作 | Action | Scope (范围) |
| :--- | :--- | :--- |
| **读取** | `dashboards:read` | `dashboards:*`, `dashboards:uid:*`, `folders:*`, `folders:uid:*` |
| **写入** | `dashboards:write` | 同上，且需要目标文件夹的读写权限 |
| **删除** | `dashboards:delete` | 同上 |
| **创建** | `dashboards:create` | 需要 `folders:*` 或 `folders:uid:*` |

> **提示**：旧版 API (`/api/dashboards/db`) 的权限逻辑与此类似，但在处理文件夹权限时需格外注意。

#### 3.3 其他辅助 API
*   **获取首页仪表盘**：
    ```http
    GET /api/dashboards/home
    ```
*   **获取仪表盘标签**：
    ```http
    GET /api/dashboards/tags
    ```
*   **仪表盘搜索**：
    *   请参考 [Folder/Dashboard Search API](/docs/grafana/latest/developer-resources/api-reference/http-api/folder_dashboard_search/)。
