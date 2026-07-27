import { request } from '@/utils/request';

export interface UserItem {
  id: number;
  username: string;
  real_name: string;
  phone: string;
  email: string;
  position: string;
  hire_date?: string; // 入职日期
  dept_id: number;
  dept_name?: string; // 部门名称（前端动态添加）
  status: number;
  create_time: string;
  update_time?: string;
  roles?: Array<{
    id: number;
    role_name: string;
  }>;
}

export interface UserListResult {
  list: UserItem[];
  total: number;
  page?: number;
  page_size?: number;
}

export interface UserListParams {
  page?: number;
  page_size?: number;
  username?: string;
  real_name?: string;
  status?: number;   // 状态筛选
  dept_id?: number;  // 部门筛选
}

export interface CreateUserParams {
  username: string;
  password: string;
  real_name?: string;
  phone?: string;
  email?: string;
  position?: string;
  hire_date?: string; // 入职日期
  dept_id?: number;
  status?: number;
  role_ids?: number[]; // 角色ID数组
}

export interface UpdateUserParams {
  id: number;
  username?: string;
  real_name?: string;
  phone?: string;
  email?: string;
  position?: string;
  hire_date?: string; // 入职日期
  dept_id?: number;
  status?: number;
  role_ids?: number[]; // 角色ID数组
}

const Api = {
  UserList: '/system/users',
  UserDetail: '/system/users/:id',
  CreateUser: '/system/users',
  UpdateUser: '/system/users',
  DeleteUser: '/system/users/:id',
  BatchStatus: '/system/users/status',
  ResetPassword: '/system/users/:id/reset-pwd',
  UnlockUser: '/system/users/:id/unlock',
};

/**
 * 获取用户列表（支持分页和搜索）
 */
export function getUserList(params?: UserListParams) {
  return request.get<UserListResult>({
    url: Api.UserList,
    params,
  });
}

/**
 * 获取用户详情
 */
export function getUserDetail(id: number) {
  return request.get<UserItem>({
    url: Api.UserDetail.replace(':id', id.toString()),
  });
}

/**
 * 创建用户
 */
export function createUser(data: CreateUserParams) {
  return request.post({
    url: Api.CreateUser,
    data,
  });
}

/**
 * 更新用户
 */
export function updateUser(data: UpdateUserParams) {
  return request.put({
    url: Api.UpdateUser,
    data,
  });
}

/**
 * 删除用户
 */
export function deleteUser(id: number) {
  return request.delete({
    url: Api.DeleteUser.replace(':id', id.toString()),
  });
}

/**
 * 批量更新用户状态
 */
export function batchUpdateUserStatus(data: { ids: number[]; status: number }) {
  return request.put({
    url: Api.BatchStatus,
    data,
  });
}

/**
 * 管理员重置用户密码
 */
export function resetUserPassword(id: number, newPassword: string) {
  return request.put({
    url: Api.ResetPassword.replace(':id', id.toString()),
    data: { new_password: newPassword },
  });
}

/**
 * 管理员解锁用户账号
 */
export function unlockUser(id: number) {
  return request.put({
    url: Api.UnlockUser.replace(':id', id.toString()),
  });
}
