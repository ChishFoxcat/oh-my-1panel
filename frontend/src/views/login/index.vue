<script setup lang="ts">
import type { CaptchaResponse, LoginResponse } from '@/api/types'
import { ArrowLeftIcon, EyeIcon, EyeOffIcon, KeyRoundIcon, MoonIcon, SunIcon, UserIcon } from '@lucide/vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { ApiError } from '@/api/request'
import { fetchCaptcha, login, loginWithMFA } from '@/api/auth'
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp'
import { Spinner } from '@/components/ui/spinner'
import { useSystemStore } from '@/stores/system'
import { useThemeStore } from '@/stores/theme'

const otpLength = 6

const router = useRouter()
const system = useSystemStore()
const theme = useThemeStore()

const step = ref<'password' | 'mfa'>('password')
const submitting = ref(false)
const showPassword = ref(false)
const captchaLoading = ref(false)
const captcha = ref<CaptchaResponse>({ captchaID: '', imagePath: '' })
const captchaRequired = ref(false)
const errorMessage = ref('')

const form = reactive({ name: '', password: '', captcha: '' })
const mfa = reactive({ sessionId: '', code: '' })

const canSubmit = computed(() => form.name.trim().length > 0 && form.password.length > 0)

onMounted(async () => {
  try {
    const info = await system.ensure()
    captchaRequired.value = info.needCaptcha
    if (captchaRequired.value) {
      await loadCaptcha()
    }
  }
  catch (thrown) {
    errorMessage.value = thrown instanceof Error ? thrown.message : '无法连接面板'
  }
})

async function loadCaptcha() {
  captchaLoading.value = true
  try {
    captcha.value = await fetchCaptcha()
  }
  catch (thrown) {
    toast.error(thrown instanceof Error ? thrown.message : '验证码加载失败')
  }
  finally {
    captchaLoading.value = false
  }
}

async function submitPassword() {
  if (!canSubmit.value || submitting.value) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const result = await login({
      name: form.name.trim(),
      password: form.password,
      captcha: form.captcha,
      captchaID: captcha.value.captchaID,
    })
    if (result.mfaRequired) {
      mfa.sessionId = result.mfaSession ?? ''
      mfa.code = ''
      step.value = 'mfa'
      return
    }
    await finishLogin(result)
  }
  catch (thrown) {
    await handleFailure(thrown)
  }
  finally {
    submitting.value = false
  }
}

async function submitMFA() {
  if (mfa.code.length < otpLength || submitting.value) {
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const result = await loginWithMFA({ sessionId: mfa.sessionId, code: mfa.code })
    await finishLogin(result)
  }
  catch (thrown) {
    if (thrown instanceof ApiError && thrown.detail === 'ErrMFA') {
      mfa.code = ''
    }
    errorMessage.value = thrown instanceof Error ? thrown.message : '验证码校验失败'
  }
  finally {
    submitting.value = false
  }
}

async function finishLogin(result: LoginResponse) {
  form.password = ''
  mfa.code = ''
  // 刷新登录态后再把地址切到根路径：登录后不再需要安全入口
  await system.markLoggedIn(result)
  toast.success('登录成功')
  await router.replace('/')
}

/** 密码错误与验证码错误时按面板语义刷新验证码输入。 */
async function handleFailure(thrown: unknown) {
  errorMessage.value = thrown instanceof Error ? thrown.message : '登录失败'
  if (!(thrown instanceof ApiError)) {
    return
  }
  if (thrown.detail === 'ErrCaptchaCode' || thrown.detail === 'ErrAuth') {
    captchaRequired.value = true
    form.captcha = ''
    await loadCaptcha()
  }
}

function backToPassword() {
  step.value = 'password'
  mfa.code = ''
  errorMessage.value = ''
}
</script>

<template>
  <div class="bg-muted/30 flex min-h-svh flex-col">
    <header class="flex items-center justify-between px-4 py-3 sm:px-6">
      <div class="text-muted-foreground flex items-center gap-2 text-sm">
        <span class="text-foreground font-medium">{{ system.panelName }}</span>
        <span class="hidden sm:inline">{{ system.info?.panelURL }}</span>
      </div>
      <Button
        variant="ghost"
        size="icon"
        :aria-label="theme.theme === 'dark' ? '切换到亮色模式' : '切换到暗色模式'"
        @click="theme.toggle()"
      >
        <MoonIcon v-if="theme.theme === 'dark'" />
        <SunIcon v-else />
      </Button>
    </header>

    <main class="flex flex-1 items-center justify-center px-4 pb-10">
      <Card class="w-full max-w-sm">
        <CardHeader>
          <CardTitle>{{ step === 'password' ? '登录' : '动态验证码' }}</CardTitle>
          <CardDescription>
            <template v-if="step === 'password'">
              使用面板账号登录，凭据由服务端加密后转发
            </template>
            <template v-else>
              请输入身份验证器中的 6 位动态验证码
            </template>
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form
            v-if="step === 'password'"
            @submit.prevent="submitPassword"
          >
            <FieldGroup>
              <Field>
                <FieldLabel for="login-name">账号</FieldLabel>
                <InputGroup>
                  <InputGroupInput
                    id="login-name"
                    v-model="form.name"
                    name="username"
                    autocomplete="username"
                    placeholder="请输入面板账号"
                    required
                    :aria-invalid="errorMessage ? true : undefined"
                  />
                  <InputGroupAddon>
                    <UserIcon />
                  </InputGroupAddon>
                </InputGroup>
              </Field>

              <Field>
                <FieldLabel for="login-password">密码</FieldLabel>
                <InputGroup>
                  <InputGroupInput
                    id="login-password"
                    v-model="form.password"
                    :type="showPassword ? 'text' : 'password'"
                    name="password"
                    autocomplete="current-password"
                    placeholder="请输入登录密码"
                    required
                    :aria-invalid="errorMessage ? true : undefined"
                  />
                  <InputGroupAddon align="inline-end">
                    <InputGroupButton
                      size="icon-xs"
                      :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                      @click="showPassword = !showPassword"
                    >
                      <EyeOffIcon v-if="showPassword" />
                      <EyeIcon v-else />
                    </InputGroupButton>
                  </InputGroupAddon>
                </InputGroup>
              </Field>

              <Field v-if="captchaRequired">
                <FieldLabel for="login-captcha">验证码</FieldLabel>
                <div class="grid grid-cols-[1fr_auto] gap-2">
                  <Input
                    id="login-captcha"
                    v-model="form.captcha"
                    name="captcha"
                    autocomplete="off"
                    placeholder="请输入计算结果"
                    required
                  />
                  <Button
                    type="button"
                    variant="outline"
                    class="h-8 px-1.5"
                    aria-label="刷新验证码"
                    :disabled="captchaLoading"
                    @click="loadCaptcha"
                  >
                    <img
                      v-if="captcha.imagePath"
                      :src="captcha.imagePath"
                      alt="验证码"
                      class="h-6 w-auto"
                    >
                    <Spinner v-else />
                  </Button>
                </div>
                <FieldDescription>点击图片可刷新验证码</FieldDescription>
              </Field>

              <Field>
                <Button
                  type="submit"
                  :disabled="submitting || !canSubmit"
                >
                  <Spinner v-if="submitting" data-icon="inline-start" />
                  <KeyRoundIcon v-else data-icon="inline-start" />
                  登录
                </Button>
              </Field>
            </FieldGroup>
          </form>

          <form
            v-else
            @submit.prevent="submitMFA"
          >
            <FieldGroup>
              <Field>
                <FieldLabel for="login-mfa">动态验证码</FieldLabel>
                <InputOTP
                  id="login-mfa"
                  v-model="mfa.code"
                  :maxlength="otpLength"
                  :disabled="submitting"
                >
                  <InputOTPGroup>
                    <InputOTPSlot
                      v-for="index in otpLength"
                      :key="index"
                      :index="index - 1"
                    />
                  </InputOTPGroup>
                </InputOTP>
                <FieldDescription>验证码每 30 秒刷新一次</FieldDescription>
              </Field>

              <Field>
                <Button
                  type="submit"
                  :disabled="submitting || mfa.code.length < otpLength"
                >
                  <Spinner v-if="submitting" data-icon="inline-start" />
                  <KeyRoundIcon v-else data-icon="inline-start" />
                  验证并登录
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  :disabled="submitting"
                  @click="backToPassword"
                >
                  <ArrowLeftIcon data-icon="inline-start" />
                  返回账号登录
                </Button>
              </Field>
            </FieldGroup>
          </form>

          <Alert
            v-if="errorMessage"
            variant="destructive"
            class="mt-4"
          >
            <AlertTitle>登录失败</AlertTitle>
            <AlertDescription>{{ errorMessage }}</AlertDescription>
          </Alert>

          <FieldError
            v-else-if="system.error"
            class="mt-4"
            :errors="[system.error]"
          />
        </CardContent>

        <CardFooter>
          <p class="text-muted-foreground text-xs">
            凭据由服务端加密后转发至面板，浏览器不保存密码。
          </p>
        </CardFooter>
      </Card>
    </main>
  </div>
</template>
