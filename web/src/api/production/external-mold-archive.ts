import { request } from '@/utils/request';

/**
 * 根据模具编号查询外来模具档案
 * @param moldCode 模具编号
 * @param excludeOrderId 排除的订单ID（编辑模式下传入当前订单ID，避免提示当前编辑的订单）
 */
export function getExternalMoldArchiveByCode(moldCode: string, excludeOrderId?: number) {
  const params: any = { mold_code: moldCode };
  if (excludeOrderId) {
    params.exclude_order_id = excludeOrderId;
  }
  return request.get({
    url: '/production/external-mold-archive/query',
    params,
  });
}

/**
 * 查询外来模具档案列表
 */
export function getExternalMoldArchiveList(params: {
  mold_code?: string;
  customer_id?: number;
  page?: number;
  page_size?: number;
}) {
  return request.get({ url: '/production/external-mold-archive/list', params });
}

/**
 * 更新外来模具档案
 */
export function updateExternalMoldArchive(id: number, data: {
  product_name: string;
  remark?: string;
}) {
  return request.put({ url: `/production/external-mold-archive/${id}`, data });
}

/**
 * 删除外来模具档案（软删除）
 */
export function deleteExternalMoldArchive(id: number) {
  return request.delete({ url: `/production/external-mold-archive/${id}` });
}
