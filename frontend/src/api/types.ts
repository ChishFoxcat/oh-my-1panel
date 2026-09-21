/** 后端统一响应信封：code 与 HTTP 状态码一致，detail 仅在错误来自面板时携带面板错误键。 */
export type Envelope<T> = {
  code: number
  message: string
  detail?: string
  data: T
}

/** 当前登录用户的最小信息。 */
export type UserInfo = {
  name: string
  role: string
  isAdmin: boolean
}

/** 面板公开信息与登录态。 */
export type SystemInfo = {
  panelName: string
  panelURL: string
  language: string
  theme: string
  needCaptcha: boolean
  logged: boolean
  user?: UserInfo
  version: string
}

/** 登录验证码。 */
export type CaptchaResponse = {
  captchaID: string
  imagePath: string
}

/** 账号密码登录请求。 */
export type LoginPayload = {
  name: string
  password: string
  captcha?: string
  captchaID?: string
}

/** 动态验证码校验请求。 */
export type MFAPayload = {
  sessionId: string
  code: string
}

/** 登录结果：需要动态验证码时返回 mfaRequired 与 mfaSession。 */
export type LoginResponse = {
  mfaRequired: boolean
  mfaSession?: string
  expiresAt?: string
  user?: UserInfo
}

/** 面板当前登录用户。 */
export type CurrentUser = {
  name: string
  role: string
  isAdmin: boolean
  mfaStatus: string
  authSource: string
  permissions: string[]
}
