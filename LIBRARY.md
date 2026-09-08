# 公开对话图书馆

聊天页左侧书架默认展开，可收起到页面左端。书架和聊天工具栏中的“记录进图书馆”会打开当前对话的快照预览；填写标题并确认后才会公开。标准版、经典版、自设版均取当前可见对话，不公开其他会话或人格设定。

访客可浏览所有公开记录及留言；发布、点赞、取消点赞和留言需要登录。用户名取自服务端验证的账号。公开记录是独立快照，继续聊天或删除私人对话不会改变它。暂未提供撤回公开记录的界面。

数据存入现有 `RAG_DB_PATH` 指向的 SQLite 数据库，首次请求自动创建 `app_library`、`app_library_likes` 和 `app_library_comments` 表。部署需要同时更新前端构建和 Go 服务。开发代理已将 `/api` 转发至 `localhost:3000`。

接口（均在 `/api/library` 下）：

- `GET /?offset=0`：每页30条公开记录，按发布时间倒序。
- `POST /`：`{title, messages: [{name, content}]}`，发布快照。
- `GET /:id`：完整记录及点赞、留言数。
- `PUT /:id/like`：`{liked: true|false}`，同账号操作幂等。
- `GET /:id/comments?offset=0`：每页30条留言，最新在前。
- `POST /:id/comments`：`{content}`，发表留言。

标题最多80字；快照最多2000条且编码后不超过2MB；留言最多1000字。内容作为纯文本显示。发布和留言沿用输入审计。

验证：在 `server` 执行 `go test ./...`；在 `client` 执行 `node scripts/library-smoke.js` 和 `npm run build`。冒烟测试使用模拟接口，不代替真实浏览器联调。
