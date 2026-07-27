import { request } from '@/utils/request';

export interface SearchResultItem {
  id: number;
  title: string;
  description: string;
  module: string;
  module_name: string;
  path: string;
}

export interface SearchResponse {
  items: SearchResultItem[];
  total: number;
}

export interface SearchParams {
  keyword: string;
  page?: number;
  page_size?: number;
}

export function globalSearch(params: SearchParams) {
  return request.get<SearchResponse>({ url: '/search', params });
}
