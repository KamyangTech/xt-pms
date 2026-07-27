import { request } from '@/utils/request';

export interface RoleItem {
  id: number;
  role_name: string;
  remark: string;
  role_type: number;   // 0=自定义 1=系统内置
  data_scope: number;  // 1=全部 2=本部门及以下 3=本部门 4=仅本人
  status: number;      // 0=禁用 1=启用
  create_time: string;
  update_time: string;
}

export interface RoleListResult {
  list: RoleItem[];
  total: number;
}

export interface CreateRoleParams {
  role_name: string;
  remark?: string;
}

export interface UpdateRoleParams {
  id: number;
  role_name: string;
  remark?: string;
  role_type?: number;
  data_scope?: number;
  status?: number;
}

const Api = {
  RoleList: '/system/roles',
  RoleDetail: '/system/roles/:id',
  CreateRole: '/system/roles',
  UpdateRole: '/system/roles',
  DeleteRole: '/system/roles/:id',
};

/**
 * 获取角色列表
 */
export function getRoleList() {
  return request.get<RoleListResult>({
    url: Api.RoleList,
  });
}

/**
 * 获取角色详情
 */
export function getRoleDetail(id: number) {
  return request.get<RoleItem>({
    url: Api.RoleDetail.replace(':id', id.toString()),
  });
}

/**
 * 创建角色
 */
export function createRole(data: CreateRoleParams) {
  return request.post({
    url: Api.CreateRole,
    data,
  });
}

/**
 * 更新角色
 */
export function updateRole(data: UpdateRoleParams) {
  return request.put({
    url: Api.UpdateRole,
    data,
  });
}

/**
 * 删除角色
 */
export function deleteRole(id: number) {
  return request.delete({
    url: Api.DeleteRole.replace(':id', id.toString()),
  });
}
