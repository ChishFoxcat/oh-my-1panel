<script setup lang="ts">
import { LogOutIcon, MoonIcon, ShieldCheckIcon, SunIcon } from '@lucide/vue'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'
import { fetchCurrentUser, logout } from '@/api/auth'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Spinner } from '@/components/ui/spinner'
import { useSystemStore } from '@/stores/system'
import { useThemeStore } from '@/stores/theme'

const router = useRouter()
const system = useSystemStore()
const theme = useThemeStore()

const loading = ref(true)
const loggingOut = ref(false)
const current = ref<{ name: string, role: string, isAdmin: boolean, mfaStatus: string, authSource: string, permissions: string[] } | null>(null)

onMounted(async () => {
  try {
    current.value = await fetchCurrentUser()
  }
  catch (thrown) {
    toast.error(thrown instanceof Error ? thrown.message : '读取用户信息失败')
  }
  finally {
    loading.value = false
  }
})

async function handleLogout() {
  loggingOut.value = true
  try {
    await logout()
    system.clear()
    toast.success('已退出登录')
    await router.replace({ name: 'login' })
  }
  catch (thrown) {
    toast.error(thrown instanceof Error ? thrown.message : '退出登录失败')
  }
  finally {
    loggingOut.value = false
  }
}
</script>

<template>
  <div class="bg-muted/30 flex min-h-svh flex-col">
    <header class="flex items-center justify-between px-4 py-3 sm:px-6">
      <div class="flex items-center gap-2 text-sm">
        <span class="font-medium">{{ system.panelName }}</span>
        <Badge variant="secondary">{{ system.info?.version }}</Badge>
      </div>
      <div class="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          :aria-label="theme.theme === 'dark' ? '切换到亮色模式' : '切换到暗色模式'"
          @click="theme.toggle()"
        >
          <MoonIcon v-if="theme.theme === 'dark'" />
          <SunIcon v-else />
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="loggingOut"
          @click="handleLogout"
        >
          <Spinner v-if="loggingOut" data-icon="inline-start" />
          <LogOutIcon v-else data-icon="inline-start" />
          退出登录
        </Button>
      </div>
    </header>

    <main class="flex flex-1 items-start justify-center px-4 py-6">
      <Card class="w-full max-w-2xl">
        <CardHeader>
          <CardTitle>已登录</CardTitle>
          <CardDescription>
            登录页已完成，后续功能页面将从这里展开。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="loading" class="flex items-center gap-2 text-sm">
            <Spinner />
            <span class="text-muted-foreground">正在读取面板用户信息…</span>
          </div>
          <div
            v-else-if="current"
            class="flex flex-col gap-3 text-sm"
          >
            <div class="flex items-center gap-2">
              <ShieldCheckIcon class="text-muted-foreground size-4" />
              <span class="font-medium">{{ current.name }}</span>
              <Badge :variant="current.isAdmin ? 'default' : 'secondary'">
                {{ current.role }}
              </Badge>
            </div>
            <Separator />
            <dl class="grid grid-cols-[7rem_1fr] gap-2">
              <dt class="text-muted-foreground">认证来源</dt>
              <dd>{{ current.authSource || '—' }}</dd>
              <dt class="text-muted-foreground">动态验证码</dt>
              <dd>{{ current.mfaStatus === 'Enable' ? '已开启' : '未开启' }}</dd>
              <dt class="text-muted-foreground">面板地址</dt>
              <dd class="truncate">{{ system.info?.panelURL }}</dd>
              <dt class="text-muted-foreground">会话到期</dt>
              <dd>{{ system.expiresAt || '—' }}</dd>
              <dt class="text-muted-foreground">权限项</dt>
              <dd>{{ current.permissions.length }} 项</dd>
            </dl>
          </div>
        </CardContent>
        <CardFooter>
          <p class="text-muted-foreground text-xs">
            当前会话保存在服务端，浏览器仅持有会话与 CSRF Cookie。
          </p>
        </CardFooter>
      </Card>
    </main>
  </div>
</template>
