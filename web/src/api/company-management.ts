import { request } from '@/utils/request';

// 公司信息接口
export interface CompanyInfo {
  id: number;
  company_name: string;
  short_name: string;
  english_name: string;
  english_short_name: string;
  logo?: string;
  create_time: string;
  update_time: string;
}

const Api = {
  CompanyInfo: '/system/company',
};

// 获取公司信息（页脚展示用，跳过 403 全局重定向防止循环）
export const getCompanyInfo = () => {
  return request.get<CompanyInfo>({
    url: Api.CompanyInfo,
  } as any);
};

export const updateCompanyInfo = (data: Partial<CompanyInfo>) => {
  return request.put({
    url: Api.CompanyInfo,
    data,
  });
};
