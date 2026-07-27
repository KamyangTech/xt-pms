import { request } from '@/utils/request';

export interface LoginParams {
  username: string;
  password: string;
  captcha_id?: string;
  captcha_val?: string;
  keep_login?: boolean;
}

export interface CaptchaResult {
  captcha_id: string;
  captcha_img: string;
}

export interface LoginResult {
  token: string;
  userInfo: {
    user_id: number;
    username: string;
    real_name: string;
    roles?: string[]; // 角色列表
  };
}

export interface UserProfile {
  id: number;
  username: string;
  real_name: string;
  phone: string;
  email: string;
  position: string;
  hire_date?: string; // 入职日期
  dept_id: number;
  status: number;
  create_time: string;
  roles?: Array<{
    id: number;
    role_name: string;
  }>;
}

export interface UpdateProfileParams {
  real_name?: string;
  phone?: string;
  email?: string;
  position?: string;
  hire_date?: string;
}

export interface ChangePasswordParams {
  old_password: string;
  new_password: string;
}

const Api = {
  Login: '/system/login',
  Captcha: '/system/captcha',
  Profile: '/system/profile', // 获取用户信息
  ProfileUpdate: '/system/profile', // 更新个人信息 PUT
  ChangePassword: '/system/profile/password', // 修改密码 POST
};

/**
 * 用户登录
 */
export function login(data: LoginParams) {
  return request.post<LoginResult>({
    url: Api.Login,
    data,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8',
    },
    requestOptions: {
      skip500Redirect: true,  // 登录异常不跳 server-error
      skip401Redirect: true,  // 登录失败不登出跳转，保留在登录页弹提示
    },
  });
}

/**
 * 获取验证码
 */
export function getCaptcha() {
  return request.get<CaptchaResult>({
    url: Api.Captcha,
  });
}

/**
 * 获取当前用户信息
 */
export function getUserProfile() {
  return request.get<UserProfile>({
    url: Api.Profile,
  });
}

/**
 * 更新个人信息
 */
export function updateProfile(data: UpdateProfileParams) {
  return request.put({
    url: Api.ProfileUpdate,
    data,
  });
}

/**
 * 修改密码
 */
export function changePassword(data: ChangePasswordParams) {
  return request.post({
    url: Api.ChangePassword,
    data,
  });
}

// 用户统计
export interface UserStatistics {
  mold_order_count: number;
  production_order_count: number;
}

/**
 * 获取当前用户的统计数据
 */
export function getUserStatistics() {
  return request.get<UserStatistics>({
    url: '/system/profile/statistics',
  });
}
