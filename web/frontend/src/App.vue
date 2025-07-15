<template>
  <div id="app">
    <el-container class="layout-container">
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <h2>Kun Gateway</h2>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
        >
          <el-menu-item index="/">
            <el-icon><Monitor /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/routes">
            <el-icon><Connection /></el-icon>
            <span>路由管理</span>
          </el-menu-item>
          <el-menu-item index="/services">
            <el-icon><Service /></el-icon>
            <span>服务发现</span>
          </el-menu-item>
          <el-menu-item index="/metrics">
            <el-icon><TrendCharts /></el-icon>
            <span>监控指标</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-container>
        <el-header class="header">
          <div class="header-content">
            <h3>{{ pageTitle }}</h3>
            <div class="header-actions">
              <el-button type="primary" size="small" @click="refreshData">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </div>
        </el-header>
        
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Connection, Service, TrendCharts, Refresh } from '@element-plus/icons-vue'

const route = useRoute()

const pageTitle = computed(() => {
  const titles = {
    '/': '仪表盘',
    '/routes': '路由管理',
    '/services': '服务发现',
    '/metrics': '监控指标'
  }
  return titles[route.path] || '仪表盘'
})

const refreshData = () => {
  // 触发全局刷新事件
  window.dispatchEvent(new CustomEvent('refresh-data'))
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  color: white;
}

.logo {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #435266;
}

.logo h2 {
  margin: 0;
  color: #fff;
}

.sidebar-menu {
  border-right: none;
  background-color: #304156;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-content h3 {
  margin: 0;
  color: #303133;
}

.main-content {
  background-color: #f5f7fa;
  padding: 20px;
}
</style> 