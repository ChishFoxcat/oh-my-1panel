import type {
  CaptchaResponse,
  CurrentUser,
  LoginPayload,
  LoginResponse,
  MFAPayload,
  SystemInfo,
} from './types'
import { request } from './request'

/** 面板信息与登录态。 */
export function fetchSystemInfo(): Promise<SystemInfo> {
  return request<SystemInfo>({ url: '/system/info', method: 'get' })
}

/** 面板当前登录用户。 */
export function fetchCurrentUser(): Promise<CurrentUser> {
  return request<CurrentUser>({ url: '/system/current', method: 'get' })
}

/** 获取登录验证码。 */
export function fetchCaptcha(): Promise<CaptchaResponse> {
  return request<CaptchaResponse>({ url: '/auth/captcha', method: 'get' })
}

/** 账号密码登录。 */
export function login(payload: LoginPayload): Promise<LoginResponse> {
  return request<LoginResponse>({ url: '/auth/login', method: 'post', data: payload })
}

/** 动态验证码二次校验。 */
export function loginWithMFA(payload: MFAPayload): Promise<LoginResponse> {
  return request<LoginResponse>({ url: '/auth/mfa', method: 'post', data: payload })
}

/** 退出登录。 */
export function logout(): Promise<null> {
  return request<null>({ url: '/auth/logout', method: 'post' })
}
