<template>
  <div class="metrics-page">
    <el-card shadow="hover">
      <template #header>
        <span>监控指标</span>
      </template>
      <div>监控指标页面正在开发中...</div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getMetrics } from '../api/metrics'

const metrics = ref({})

const loadMetrics = async () => {
  try {
    metrics.value = await getMetrics()
  } catch (error) {
    console.error('加载监控指标失败:', error)
  }
}

onMounted(() => {
  loadMetrics()
})
</script>

<style scoped>
.metrics-page {
  padding: 20px;
}
</style> 