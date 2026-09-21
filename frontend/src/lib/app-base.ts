/**
 * 应用资源基址：由后端注入的 <base href> 决定，固定为 /。
 * 前端产物按站点根路径构建，安全入口不参与资源与接口路径。
 */
export function appBase(): string {
  const href = document.querySelector('base')?.getAttribute('href') ?? '/'
  return href.endsWith('/') ? href : `${href}/`
}

/**
 * 安全入口挂载点（如 /chish），未启用入口时为空串。
 *
 * 入口只作为"进入凭证"：访问该路径会拿到 base64(入口) 的 Cookie，登录后地址栏不再带它。
 * 路由需要把该路径也登记为首页，避免被兜底路由重定向掉。
 */
export function entranceMount(): string {
  const value = document.querySelector('meta[name="omop-entrance"]')?.getAttribute('content') ?? ''
  return value.replace(/\/$/, '')
}
