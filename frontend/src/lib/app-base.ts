/**
 * 应用基址：由后端注入的 <base href> 决定。
 * 启用安全入口时后端会写 <base href="/入口/">，此时前端路由与接口都落在入口路径之下；
 * 开发服务器不注入该标签，基址回落为 /。
 */
export function appBase(): string {
  const href = document.querySelector('base')?.getAttribute('href') ?? '/'
  return href.endsWith('/') ? href : `${href}/`
}
