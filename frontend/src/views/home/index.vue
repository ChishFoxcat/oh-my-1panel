<script setup lang="ts">
import LoginView from '@/views/login/index.vue'
import OverviewView from '@/views/overview/index.vue'
import { useSystemStore } from '@/stores/system'

// 首页：未登录渲染登录表单，已登录渲染概览，二者共用安全入口这一个地址。
const system = useSystemStore()
</script>

<template>
  <OverviewView v-if="system.logged" />
  <LoginView v-else />
</template>
