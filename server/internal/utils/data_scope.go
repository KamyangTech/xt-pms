package utils

// DataScope 数据权限常量
const (
	DataScopeAll        int8 = 1 // 全部数据
	DataScopeDeptAndChild int8 = 2 // 本部门及以下
	DataScopeDept       int8 = 3 // 本部门
	DataScopeSelf        int8 = 4 // 仅本人
)

// GetUserMaxDataScope 获取用户的最大数据权限范围（取最宽松的）
// 如果有多个角色，取 data_scope 值最小的（1 > 2 > 3 > 4）
func GetUserMaxDataScope(roleDataScopes []int8) int8 {
	if len(roleDataScopes) == 0 {
		return DataScopeAll // 无角色默认全部数据
	}

	minScope := roleDataScopes[0]
	for _, scope := range roleDataScopes[1:] {
		if scope < minScope {
			minScope = scope
		}
	}
	return minScope
}
