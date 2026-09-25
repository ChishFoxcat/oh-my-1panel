# AGENTS.md

本仓库的硬性约定。改动前先读这里。

## 禁止直接修改 shadcn 组件

`frontend/src/shadcn/components/ui/**` 下的文件全部由 shadcn-vue registry 生成，**禁止直接编辑**。`frontend/src/shadcn/lib/utils.ts` 等同类生成物同理。

原因：`shadcn-vue add <component>`（尤其带 `--overwrite`）会用 registry 版本整体覆盖这些文件。本地改动会**静默丢失**——不报错、不提示，只在某次重装后表现为行为或样式回退，极难定位。

需要定制时按下表处理：

| 需求 | 做法 |
| --- | --- |
| 颜色、尺寸、间距等外观 | 通过 `class` prop 传入 tailwind 类；组件内部用 `cn()` 合并，冲突项由 tailwind-merge 解析 |
| 主题色、圆角等 design token | 改 `frontend/src/shadcn/styles/globals.css` 里的 CSS 变量（`--primary`、`--radius` 等）。该文件由 CLI 管理，改动限制在变量值上，不要重排或删除既有内容 |
| 结构、行为、额外动画 | 在 `frontend/src/components/` 下封装自己的组件包住它；需要触碰组件内部元素时用 scoped 样式 + `:deep()` |
| 动画关键帧 | 写在引用方的 scoped `@keyframes` 里，或项目自有样式文件里，不要写进 registry 生成的文件 |

参照实现：登录页加载态的循环进度条。`frontend/src/components/views/auth/LoginLoading.vue` 不传 `model-value`，用 scoped `:deep([data-slot='progress-indicator'])` 加 `@keyframes` 接管指示条的宽度与 `transform`；`Progress.vue` 保持 registry 原样。（CSS 动画在层叠上优先于内联样式，所以外部样式足以覆盖组件内部的 `translateX(...)` 内联值。）

升级或新增组件前，先确认目标路径没有本地改动：

```bash
git diff -- frontend/src/shadcn/components/ui
```

无输出才可安全执行 `shadcn-vue add`；有改动先按上表把定制外移，禁止无脑 `--overwrite`。

## 目录约定

| 内容 | 位置 |
| --- | --- |
| 页面视图 | `frontend/src/views/<域>/` |
| 视图内的页面组件 | `frontend/src/components/views/<域>/` |
| 视图级类型 | `frontend/src/types/views/` |
| 页面组件类型 | `frontend/src/types/components/views/` |
| i18n 文案 | `frontend/src/i18n/lang/<locale>/<namespace>.ts`，文件名即命名空间（键路径为 `<namespace>.<分组>.<字段>`） |

## 代码约定

- 每个新建源文件保留 fileheader 模板（配置见 `workspace.code-workspace`）：`Copyright` / `Date` / `LastEditTime` / `FilePath`。
- 面向用户的文案一律走 i18n，不在模板里硬编码，品牌名同样集中登记在 locale 文件里。
- 样式优先用 tailwind 工具类表达，组件内不写与工具类重复的手写样式。
- 引入 registry 组件时按需 `add`，并保持 `frontend/components.json` 的 `style` / `baseColor` 不变，避免全项目样式漂移。

## 提交约定

- 提交信息用中文，格式 `type: 描述`（`feat` / `fix` / `chore` / `refactor` / `docs` / `style`）。
- 涉及 issue 时在正文引用（`#123`），完全解决则一并 close。
- `frontend/.gitignore` 之外不要提交 `.env*` 等含密钥的文件。
