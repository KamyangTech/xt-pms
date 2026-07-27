# Data Seeds

存放功能数据初始化 SQL 文件，用于将某个功能的数据恢复至初始状态。

## 命名规范

`<序号>_<功能描述>.sql`

例如：
- `001_dict_status_flow_seed.sql` - 字典状态流数据初始化
- `002_production_order_status_seed.sql` - 生产订单状态种子数据
- `003_menu_permission_seed.sql` - 系统菜单与权限种子数据

## 使用方式

通过项目管理工具或手动执行对应的 SQL 文件来恢复特定功能的初始数据。

## 增量迁移

**版本更新时的增量 SQL 变更**请放入 `migrations/updates/` 目录，命名格式
`vX.Y.Z.sql`。系统每次启动时会自动按版本号顺序执行未跑过的迁移文件。

详见 `server/internal/service/migration_service.go` 顶部注释。
