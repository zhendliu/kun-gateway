<template>
  <el-container class="app-container">
    <el-aside width="200px">
      <el-menu
        :default-active="$route.path"
        class="el-menu-vertical"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataBoard /></el-icon>
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
        <el-menu-item index="/certificates">
          <el-icon><Lock /></el-icon>
          <span>证书管理</span>
        </el-menu-item>
        <el-menu-item index="/metrics">
          <el-icon><TrendCharts /></el-icon>
          <span>监控指标</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup>
import { DataBoard, Connection, Service, Lock, TrendCharts } from '@element-plus/icons-vue'
</script>

<style>
.app-container {
  height: 100vh;
}

.el-menu-vertical {
  height: 100%;
  border-right: none;
}

.el-main {
  padding: 0;
  background-color: #f5f5f5;
}
</style> 