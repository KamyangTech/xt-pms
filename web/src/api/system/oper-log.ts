import { request } from '@/utils/request';

export interface OperLogItem {
  id: number;
  title: string;
  oper_type: number;
  business_type: string;
  method: string;
  request_url: string;
  request_method: string;
  oper_name: string;
  oper_ip: string;
  request_params?: string;
  json_result?: string;
  status: number;
  error_msg?: string;
  cost_time: number;
  create_time: string;
}

export interface OperLogListResult {
  list: OperLogItem[];
  total: number;
}

export interface OperLogListParams {
  page?: number;
  page_size?: number;
  title?: string;
  oper_name?: string;
  oper_type?: number;
  status?: number;
  business_type?: string;
  start_time?: string;
  end_time?: string;
}

const Api = {
  OperLogList: '/system/oper-logs',
  OperLogDetail: '/system/oper-logs/:id',
  OperLogExport: '/system/oper-log/export',
};

export function getOperLogList(params?: OperLogListParams): Promise<OperLogListResult> {
  return request.get<OperLogListResult>({
    url: Api.OperLogList,
    params,
  });
}

export function getOperLogDetail(id: number): Promise<OperLogItem> {
  return request.get<OperLogItem>({
    url: Api.OperLogDetail.replace(':id', id.toString()),
  });
}

export function getOperLogExportUrl(params: OperLogListParams): string {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, val]) => {
    if (val !== undefined && val !== '' && val !== null) {
      query.append(key, String(val));
    }
  });
  return Api.OperLogExport + '?' + query.toString();
}
