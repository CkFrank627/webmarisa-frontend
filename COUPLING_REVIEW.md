# 耦合性检查报告（webmarisa-frontend）

## 结论（TL;DR）
当前项目存在**中到高程度耦合**，主要集中在：
1. 前端同时维护两套调用栈（`fetch` + `Axios Api`）且回包结构假设不一致；
2. 前端业务核心直接依赖浏览器存储键名与远程 IP 服务；
3. 后端 Service 层耦合了分词实现与具体阈值算法，并混入 `goto` 控制流；
4. Repository 层使用 `map[string]interface{}` 作为入参并强制类型断言；
5. 路由装配直接依赖全局单例数据库与框架级注入，测试替换成本高。

---

## 发现的主要耦合点

### 1) 前端 Core 与接口层耦合方式不一致（高）
- `Core.replyStandard` 直接 `fetch('/Reply')`，但 `Core.reply` 的兜底又走 `Api.fecthMemory()`（Axios 封装）。
- 同一业务路径存在两套 HTTP 客户端与两套响应假设，导致维护与排查复杂度提升。

**证据位置**
- `client/src/core/index.ts` 中 `replyStandard` 与 `reply` 的双实现。
- `client/src/api/index.ts` 与 `client/src/api/v1/index.ts` 中 Axios 封装。

**风险**
- 一处改接口协议，另一处忘改即出现“部分可用/部分异常”。
- 认证头、错误处理、超时策略无法统一。

**建议**
- 统一到单一 API Gateway（推荐全部走 `Api`，或全部走 `fetch` 封装层）。
- 统一响应 DTO 与错误对象，避免业务层分支兼容旧逻辑。

### 2) Core 对本地存储键名与鉴权协议硬编码（中高）
- `getToken()` 同时扫描多个键：`wm_token`/`token`/`marisa_token`/`auth_token`。
- `authHeaders()` 直接拼接 `Authorization: Bearer`。

**风险**
- 登录模块一旦调整 key 命名，核心逻辑需同步改动。
- 多来源 token 规则分散，行为不可预测（优先级隐式）。

**建议**
- 抽离 `AuthStore`（唯一 token 来源）并在 API 层统一注入 header。
- 将 token key 定义为常量并集中管理。

### 3) Core 的“教学”流程与第三方 IP 服务强耦合（中）
- `teach()` 调用 `Tools.getIp()`，而 `getIp()` 强依赖 `https://ipv4.icanhazip.com/`。

**风险**
- 外网波动会直接阻断 teach 功能。
- 业务逻辑依赖基础设施能力，难以离线/内网部署。

**建议**
- 由后端从请求上下文提取客户端 IP；前端不再主动获取公网 IP。
- 若必须前端采集，至少提供降级策略（失败时传空或匿名标识）。

### 4) Service 层同时耦合分词实现、匹配阈值、数据访问（高）
- `memoriseService` 内部直接调用 `segment.Init().Cut()`、字符串切分与阈值判断（0.6 / 0.4）并直接访问 Repo。
- 业务策略（匹配算法）与基础能力（分词器）混在同一类。

**风险**
- 算法升级或替换分词器需要改动同一模块，多处连锁影响。
- 单元测试需要真实分词/数据环境，mock 粒度粗。

**建议**
- 引入 `Tokenizer`、`Matcher` 接口并注入到 Service。
- 将阈值配置化（配置文件/环境变量），避免硬编码。

### 5) Service 中 `goto` 与循环结构导致隐藏逻辑耦合（中）
- `Add()` / `Reply()` 通过 `goto DATA` 跳出多层循环。
- 在 `Add()` 中，`for keyword` 循环里 `if/else` 都会 `goto`，导致只处理很早分支，逻辑可读性与可演化性较差。

**风险**
- 稍作修改就可能破坏控制流，形成“结构性耦合”（必须理解跳转路径才能改）。

**建议**
- 改为早返回或提取纯函数 `matchKeywords()`，显式返回匹配结果。

### 6) Repository 入参使用 `map[string]interface{}`（中）
- `AddMemory(data map[string]interface{})` 在内部进行字符串断言。

**风险**
- 调用方与仓储层靠字符串键隐式耦合，编译期无法发现字段错误。

**建议**
- 改为强类型 DTO/实体入参（如 `Models.Memorise` 或 `CreateMemoryInput`）。

### 7) 路由装配对全局单例数据库耦合（中）
- `Routes.Configure` 里直接 `Datasource.GetInstace().GetMysqlDB()` 并组装 Repo/Service。
- `main` 依赖“先连 DB 再 Configure”的顺序约束。

**风险**
- 初始化顺序一旦变化会引发运行时问题。
- 无法轻易在测试时替换为内存仓储。

**建议**
- 在 `main` 显式构建依赖图并传入 `Configure(app, deps)`。
- 避免框架层直接触达全局单例。

---

## 优先级建议（按投入产出比）
1. **先统一前端 API 调用栈**（收益最高，影响面可控）。
2. **替换 Repository 的 map 入参为强类型**（快速降低隐式耦合）。
3. **把分词/匹配从 Service 拆分成接口**（为后续算法演进铺路）。
4. **去掉前端公网 IP 依赖**（提升稳定性）。

## 快速验收标准
- 前端仅有一个 HTTP 访问入口；
- `Core` 不再关心 token 多键 fallback；
- `AddMemory` 无 `map[string]interface{}`；
- `Service` 不再直接 `segment.Init()`；
- 教学功能在无外网环境可用。
