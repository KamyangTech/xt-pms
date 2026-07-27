import { request } from '@/utils/request';

export interface ReceivingAccount {
  id: number;
  account_name: string;
  account_no: string;
  bank_name: string;
  account_type: string;
  sort: number;
  status: number;
  remark: string;
  created_by: number;
  create_time: string;
  update_time: string;
}

export interface ReceivingAccountListResult {
  list: ReceivingAccount[];
  total: number;
}

export interface CreateReceivingAccountParams {
  account_name: string;
  account_no: string;
  bank_name?: string;
  account_type?: string;
  sort?: number;
  status?: number;
  remark?: string;
}

export interface UpdateReceivingAccountParams {
  id: number;
  account_name: string;
  account_no: string;
  bank_name?: string;
  account_type?: string;
  sort?: number;
  status?: number;
  remark?: string;
}

const Api = {
  List: '/system/receiving-accounts',
  Detail: '/system/receiving-accounts/:id',
  Create: '/system/receiving-accounts',
  Update: '/system/receiving-accounts',
  Delete: '/system/receiving-accounts/:id',
  BatchDelete: '/system/receiving-accounts/batch-delete',
  UpdateStatus: '/system/receiving-accounts/:id/status',
};

export function getReceivingAccountList(params?: any) {
  return request.get<ReceivingAccountListResult>({
    url: Api.List,
    params,
  });
}

export function getReceivingAccountDetail(id: number) {
  return request.get<ReceivingAccount>({
    url: Api.Detail.replace(':id', id.toString()),
  });
}

export function createReceivingAccount(data: CreateReceivingAccountParams) {
  return request.post({
    url: Api.Create,
    data,
  });
}

export function updateReceivingAccount(data: UpdateReceivingAccountParams) {
  return request.put({
    url: Api.Update,
    data,
  });
}

export function deleteReceivingAccount(id: number) {
  return request.delete({
    url: Api.Delete.replace(':id', id.toString()),
  });
}

export function batchDeleteReceivingAccounts(ids: number[]) {
  return request.post({
    url: Api.BatchDelete,
    data: { ids },
  });
}

export function updateReceivingAccountStatus(id: number, status: number) {
  return request.put({
    url: Api.UpdateStatus.replace(':id', id.toString()),
    data: { status },
  });
}
