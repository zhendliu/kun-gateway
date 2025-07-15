<template>
  <div class="services-page">
    <el-row :gutter="20">
      <!-- 服务列表 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>K8s服务列表</span>
              <el-button @click="loadServices" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-table :data="servicesList" style="width: 100%" v-loading="loading">
            <!-- 调试信息 -->
            <template #empty>
              <div>暂无数据 (总数: {{ servicesList.length }})</div>
            </template>
            <el-table-column prop="Name" label="服务名称" />
            <el-table-column prop="Namespace" label="命名空间" />
            <el-table-column prop="ClusterIP" label="ClusterIP" />
            <el-table-column label="端口" width="150">
              <template #default="scope">
                <div v-if="scope.row.Ports && scope.row.Ports.length > 0">
                  <el-tag v-for="port in scope.row.Ports" :key="port.Name || port.Port" size="small" style="margin: 2px;">
                    {{ port.Port }}{{ port.Name ? ` (${port.Name})` : '' }}
                  </el-tag>
                </div>
                <span v-else style="color: #999;">无端口</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="scope">
                <el-tag type="success" size="small">正常</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 端点列表 -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>端点列表</span>
              <el-button @click="loadEndpoints" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-table :data="endpointsList" style="width: 100%" v-loading="loading">
            <!-- 调试信息 -->
            <template #empty>
              <div>暂无数据 (总数: {{ endpointsList.length }})</div>
            </template>
            <el-table-column prop="ServiceName" label="服务名称" />
            <el-table-column prop="Namespace" label="命名空间" />
            <el-table-column label="地址" width="150">
              <template #default="scope">
                <div v-for="addr in scope.row.Addresses" :key="addr">
                  <el-tag size="small">{{ addr }}</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="端口" width="80">
              <template #default="scope">
                <el-tag v-for="port in scope.row.Ports" :key="port" size="small">
                  {{ port }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.Ready ? 'success' : 'danger'" size="small">
                  {{ scope.row.Ready ? '就绪' : '未就绪' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 服务详情 -->
    <el-card shadow="hover" style="margin-top: 20px;">
      <template #header>
        <span>服务详情</span>
      </template>
      
      <el-descriptions :column="3" border>
        <el-descriptions-item label="总服务数">{{ servicesList.length }}</el-descriptions-item>
        <el-descriptions-item label="总端点数">{{ endpointsList.length }}</el-descriptions-item>
        <el-descriptions-item label="就绪端点数">{{ readyEndpointsCount }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { getServices, getEndpoints } from '../api/metrics'
import { ElMessage } from 'element-plus'

const servicesList = ref([])
const endpointsList = ref([])
const loading = ref(false)

// 计算就绪端点数量
const readyEndpointsCount = computed(() => {
  return endpointsList.value.filter(ep => ep.Ready).length
})

// 加载服务列表
const loadServices = async () => {
  loading.value = true
  try {
    const response = await getServices()
    console.log('服务API响应:', response)
    
    // 确保我们获取到正确的数据结构
    if (response && response.services) {
      const servicesData = response.services
      console.log('服务数据:', servicesData)
      
      // 详细检查每个服务的端口信息
      Object.keys(servicesData).forEach(key => {
        const service = servicesData[key]
        console.log(`服务 ${key}:`, {
          Name: service.Name,
          Namespace: service.Namespace,
          ClusterIP: service.ClusterIP,
          Ports: service.Ports,
          PortsType: typeof service.Ports,
          PortsLength: service.Ports ? service.Ports.length : 'undefined'
        })
      })
      
      servicesList.value = Object.keys(servicesData).map(key => ({
        key,
        ...servicesData[key]
      }))
      console.log('处理后的服务列表:', servicesList.value)
    } else {
      console.error('服务数据结构不正确:', response)
      servicesList.value = []
    }
  } catch (error) {
    console.error('加载服务失败:', error)
    ElMessage.error('加载服务失败')
    servicesList.value = []
  } finally {
    loading.value = false
  }
}

// 加载端点列表
const loadEndpoints = async () => {
  loading.value = true
  try {
    const response = await getEndpoints()
    console.log('端点API响应:', response)
    
    if (response && response.endpoints) {
      const endpointsData = response.endpoints
      console.log('端点数据:', endpointsData)
      endpointsList.value = Object.keys(endpointsData).map(key => ({
        key,
        ...endpointsData[key]
      }))
      console.log('处理后的端点列表:', endpointsList.value)
    } else {
      console.error('端点数据结构不正确:', response)
      endpointsList.value = []
    }
  } catch (error) {
    console.error('加载端点失败:', error)
    ElMessage.error('加载端点失败')
    endpointsList.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadServices()
  loadEndpoints()
})
</script>

<style scoped>
.services-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style> 