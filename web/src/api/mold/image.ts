/**
 * 模具图片 API（关联管理，上传走统一接口）
 */
import { request } from '@/utils/request';

// 模具图片类型（继承自统一 File，前端使用相同的字段名）
export interface MoldFile {
  id: string;
  url: string;
  file_name: string;
  file_size: number;
  file_type: string;
  is_primary: number;
  sort: number;
  create_time: string;
}

export function getMoldImages(moldId: number): Promise<MoldFile[]> {
  return request.get({
    url: `/mold/${moldId}/images`,
  });
}

// 关联已上传的文件到模具
export function attachMoldFiles(moldId: number, fileIds: string[]) {
  return request.post({
    url: `/mold/${moldId}/images/attach`,
    data: { file_ids: fileIds },
  });
}

export function deleteMoldImage(imageId: string) {
  return request.delete({
    url: `/mold/images/${imageId}`,
  });
}

export function setPrimaryImage(moldId: number, imageId: string) {
  return request.put({
    url: `/mold/${moldId}/images/${imageId}/primary`,
  });
}
