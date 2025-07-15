<template>
  <div class="routes-page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>路由管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            创建路由
          </el-button>
        </div>
      </template>
      
      <el-table :data="routes" style="width: 100%" v-loading="loading">
        <el-table-column prop="domain" label="域名" />
        <el-table-column prop="path" label="路径" />
        <el-table-column prop="service" label="目标服务" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="weight" label="权重" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'warning'">
              {{ scope.row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button size="small" @click="editRoute(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteRoute(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑路由对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRoute ? '编辑路由' : '创建路由'"
      width="600px"
    >
      <el-form :model="routeForm" label-width="100px">
        <el-form-item label="域名">
          <el-input v-model="routeForm.domain" placeholder="例如: example.com" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="routeForm.path" placeholder="例如: /api" />
        </el-form-item>
        <el-form-item label="目标服务">
          <el-select v-model="routeForm.service" placeholder="选择服务" style="width: 100%">
            <el-option
              v-for="service in services"
              :key="service.key"
              :label="service.key"
              :value="service.key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="routeForm.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="权重">
          <el-input-number v-model="routeForm.weight" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="routeForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="saveRoute">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { getRoutes, createRoute, updateRoute, deleteRoute as deleteRouteApi, getServices } from '../api/metrics'
import { ElMessage, ElMessageBox } from 'element-plus'

const routes = ref([])
const services = ref([])
const loading = ref(false)
const showCreateDialog = ref(false)
const editingRoute = ref(null)

const routeForm = ref({
  domain: '',
  path: '',
  service: '',
  port: 80,
  weight: 100,
  enabled: true
})

// 加载路由列表
const loadRoutes = async () => {
  loading.value = true
  try {
    routes.value = await getRoutes()
  } catch (error) {
    ElMessage.error('加载路由失败')
  } finally {
    loading.value = false
  }
}

// 加载服务列表
const loadServices = async () => {
  try {
    const servicesData = await getServices()
    services.value = Object.keys(servicesData).map(key => ({
      key,
      ...servicesData[key]
    }))
  } catch (error) {
    ElMessage.error('加载服务失败')
  }
}

// 编辑路由
const editRoute = (route) => {
  editingRoute.value = route
  routeForm.value = { ...route }
  showCreateDialog.value = true
}

// 删除路由
const deleteRoute = async (route) => {
  try {
    await ElMessageBox.confirm('确定要删除这个路由吗？', '确认删除', {
      type: 'warning'
    })
    
    await deleteRouteApi(route.id)
    ElMessage.success('删除成功')
    loadRoutes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 保存路由
const saveRoute = async () => {
  try {
    if (editingRoute.value) {
      await updateRoute(editingRoute.value.id, routeForm.value)
      ElMessage.success('更新成功')
    } else {
      await createRoute(routeForm.value)
      ElMessage.success('创建成功')
    }
    
    showCreateDialog.value = false
    editingRoute.value = null
    routeForm.value = {
      domain: '',
      path: '',
      service: '',
      port: 80,
      weight: 100,
      enabled: true
    }
    loadRoutes()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  loadRoutes()
  loadServices()
})
</script>

<style scoped>
.routes-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style> 