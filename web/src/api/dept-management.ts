import { request } from '@/utils/request';

export interface DeptItem {
  id: number;
  dept_name: string;
  parent_id: number;
  sort: number;
  leader: string;
  phone: string;
  status: number;
  create_time: string;
  update_time: string;
  children?: DeptItem[];
}

export interface DeptListResult {
  list: DeptItem[];
  total: number;
}

export interface CreateDeptParams {
  dept_name: string;
  parent_id?: number;
  sort?: number;
  leader?: string;
  phone?: string;
  status?: number;
}

export interface UpdateDeptParams {
  id: number;
  dept_name: string;
  parent_id?: number;
  sort?: number;
  leader?: string;
  phone?: string;
  status?: number;
}

const Api = {
  DeptList: '/system/depts',
  DeptDetail: '/system/depts/:id',
  CreateDept: '/system/depts',
  UpdateDept: '/system/depts',
  DeleteDept: '/system/depts/:id',
};

/**
 * 获取部门列表
 */
export function getDeptList() {
  return request.get<DeptListResult>({
    url: Api.DeptList,
  });
}

/**
 * 获取部门详情
 */
export function getDeptDetail(id: number) {
  return request.get<DeptItem>({
    url: Api.DeptDetail.replace(':id', id.toString()),
  });
}

/**
 * 创建部门
 */
export function createDept(data: CreateDeptParams) {
  return request.post({
    url: Api.CreateDept,
    data,
  });
}

/**
 * 更新部门
 */
export function updateDept(data: UpdateDeptParams) {
  return request.put({
    url: Api.UpdateDept,
    data,
  });
}

/**
 * 更新部门状态
 */
export function updateDeptStatus(id: number, status: number) {
  return request.put({
    url: `/system/depts/${id}/status`,
    data: { status },
  });
}

/**
 * 删除部门
 */
export function deleteDept(id: number) {
  return request.delete({
    url: Api.DeleteDept.replace(':id', id.toString()),
  });
}
