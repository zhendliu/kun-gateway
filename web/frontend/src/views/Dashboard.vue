<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <!-- 统计卡片 -->
      <el-col :span="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="24">
                <component :is="stat.icon" />
              </el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stat.value }}</div>
              <div class="stat-title">{{ stat.title }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 请求趋势图 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>请求趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="requestChartOption" style="height: 300px;" />
          </div>
        </el-card>
      </el-col>

      <!-- 状态码分布 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>状态码分布</span>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="statusChartOption" style="height: 300px;" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 活跃路由 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>活跃路由</span>
            </div>
          </template>
          <el-table :data="activeRoutes" style="width: 100%">
            <el-table-column prop="domain" label="域名" />
            <el-table-column prop="path" label="路径" />
            <el-table-column prop="requests" label="请求数" width="100" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.status === 'active' ? 'success' : 'warning'">
                  {{ scope.row.status === 'active' ? '活跃' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 系统状态 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>系统状态</span>
            </div>
          </template>
          <div class="system-status">
            <div class="status-item" v-for="status in systemStatus" :key="status.name">
              <div class="status-label">{{ status.name }}</div>
              <div class="status-value">
                <el-tag :type="status.status === 'healthy' ? 'success' : 'danger'">
                  {{ status.status === 'healthy' ? '正常' : '异常' }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { 
  Monitor, 
  Connection, 
  Service, 
  TrendCharts 
} from '@element-plus/icons-vue'
import { getMetrics } from '../api/metrics'

use([CanvasRenderer, LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

// 统计数据
const stats = ref([
  { title: '总请求数', value: '0', icon: 'Monitor', color: '#409EFF' },
  { title: '活跃连接', value: '0', icon: 'Connection', color: '#67C23A' },
  { title: '服务数量', value: '0', icon: 'Service', color: '#E6A23C' },
  { title: '路由数量', value: '0', icon: 'TrendCharts', color: '#F56C6C' }
])

// 活跃路由
const activeRoutes = ref([])

// 系统状态
const systemStatus = ref([
  { name: '数据面', status: 'healthy' },
  { name: '控制面', status: 'healthy' },
  { name: 'K8s连接', status: 'healthy' }
])

// 请求趋势图配置
const requestChartOption = ref({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: [] },
  yAxis: { type: 'value' },
  series: [{
    data: [],
    type: 'line',
    smooth: true
  }]
})

// 状态码分布图配置
const statusChartOption = ref({
  tooltip: { trigger: 'item' },
  series: [{
    type: 'pie',
    radius: '50%',
    data: [
      { value: 0, name: '2xx' },
      { value: 0, name: '3xx' },
      { value: 0, name: '4xx' },
      { value: 0, name: '5xx' }
    ]
  }]
})

let refreshTimer = null

// 加载数据
const loadData = async () => {
  try {
    const metrics = await getMetrics()
    
    // 更新统计数据
    stats.value[0].value = metrics.total_requests || 0
    stats.value[1].value = metrics.connection_count || 0
    stats.value[2].value = metrics.services_count || 0
    stats.value[3].value = metrics.routes_count || 0
    
    // 更新状态码分布
    if (metrics.status_codes) {
      const statusData = [
        { value: metrics.status_codes['2xx'] || 0, name: '2xx' },
        { value: metrics.status_codes['3xx'] || 0, name: '3xx' },
        { value: metrics.status_codes['4xx'] || 0, name: '4xx' },
        { value: metrics.status_codes['5xx'] || 0, name: '5xx' }
      ]
      statusChartOption.value.series[0].data = statusData
    }
    
  } catch (error) {
    console.error('加载数据失败:', error)
  }
}

// 开始定时刷新
const startRefresh = () => {
  refreshTimer = setInterval(loadData, 5000) // 每5秒刷新一次
}

// 停止定时刷新
const stopRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  loadData()
  startRefresh()
})

onUnmounted(() => {
  stopRefresh()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stat-card {
  margin-bottom: 20px;
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-right: 15px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-title {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  width: 100%;
}

.system-status {
  padding: 10px 0;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.status-item:last-child {
  border-bottom: none;
}

.status-label {
  font-size: 14px;
  color: #606266;
}

.status-value {
  font-size: 14px;
}
</style> 