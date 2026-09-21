/**
 * 应用基址：由后端注入的 <base href> 决定。
 * 启用安全入口时后端会写 <base href="/入口/">，此时前端资源与接口都落在入口路径之下；
 * 开发服务器不注入该标签，基址回落为 /。
 */
export function appBase(): string {
  const href = document.querySelector('base')?.getAttribute('href') ?? '/'
  return href.endsWith('/') ? href : `${href}/`
}

/**
 * 路由挂载前缀（不带结尾斜杠），例如 /chish；未启用安全入口时为空串。
 *
 * 安全入口由路由路径承载、而不是交给 vue-router 的 history base：
 * vue-router 会把 base 规范化成不带尾斜杠的形式，并在位于 base 根时把地址写成
 * base + '/'，于是访问 /chish 会被改成 /chish/。用空 base + 带前缀的路由路径
 * 才能让地址栏保持原样。
 */
export function routeMount(): string {
  return appBase().replace(/\/$/, '')
}
