import { request } from '@/utils/request';

export interface MenuItem {
  id: number;
  parent_id: number;
  menu_name: string;
  menu_type: number; // 1=目录 2=菜单 3=按钮
  permission: string;
  path: string;
  component: string;
  icon: string;
  sort: number;
  status: number;
  create_time: string;
  update_time: string;
  children?: MenuItem[];
}

export interface CreateMenuParams {
  parent_id?: number;
  menu_name: string;
  menu_type?: number;
  permission?: string;
  path?: string;
  component?: string;
  icon?: string;
  sort?: number;
  status?: number;
}

export interface UpdateMenuParams extends CreateMenuParams {
  id: number;
}

export interface MenuListResult {
  list: MenuItem[];
  total: number;
}

export interface RoleMenuBindParams {
  role_id: number;
  menu_ids: number[];
}

export interface UserPermissionsResult {
  permissions: string[];
  menus: MenuItem[];
}

const Api = {
  MenuTree: '/system/menus',
  MenuDetail: '/system/menus/:id',
  CreateMenu: '/system/menus',
  UpdateMenu: '/system/menus',
  DeleteMenu: '/system/menus/:id',
  RoleMenuTree: '/system/role-menus/tree',      // 角色分配用
  RoleMenus: '/system/roles/:id/menus',
  AssignRoleMenus: '/system/role-menus',
  UserPermissions: '/system/user/permissions',
};

/** 获取菜单树 */
export function getMenuTree() {
  return request.get<MenuListResult>({ url: Api.MenuTree });
}

/** 获取菜单详情 */
export function getMenuDetail(id: number) {
  return request.get<MenuItem>({ url: Api.MenuDetail.replace(':id', id.toString()) });
}

/** 创建菜单 */
export function createMenu(data: CreateMenuParams) {
  return request.post({ url: Api.CreateMenu, data });
}

/** 更新菜单 */
export function updateMenu(data: UpdateMenuParams) {
  return request.put({ url: Api.UpdateMenu, data });
}

/** 删除菜单 */
export function deleteMenu(id: number) {
  return request.delete({ url: Api.DeleteMenu.replace(':id', id.toString()) });
}

/** 获取菜单树（角色权限分配用） */
export function getRoleMenuTree() {
  return request.get<MenuListResult>({ url: Api.RoleMenuTree });
}

/** 获取角色已分配的菜单 ID 列表 */
export function getRoleMenus(roleId: number) {
  return request.get<{ menu_ids: number[] }>({ url: Api.RoleMenus.replace(':id', roleId.toString()) });
}

/** 为角色分配菜单 */
export function assignRoleMenus(data: RoleMenuBindParams) {
  return request.post({ url: Api.AssignRoleMenus, data });
}

/** 获取当前用户的权限和菜单 */
export function getUserPermissions() {
  return request.get<UserPermissionsResult>({ url: Api.UserPermissions });
}
