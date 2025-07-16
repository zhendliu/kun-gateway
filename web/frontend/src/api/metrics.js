import request from "@/utils/request";

// 创建API请求实例，使用统一的request工具
const api = request;

// 获取监控指标
export const getMetrics = async () => {
  try {
    const response = await api.get("/api/v1/metrics");
    return response.metrics || {};
  } catch (error) {
    console.error("获取监控指标失败:", error);
    return {};
  }
};

// 获取域名维度监控指标
export const getDomainMetrics = async (domain = "") => {
  try {
    const params = domain ? { domain } : {};
    const response = await api.get("/api/v1/metrics/domains", { params });
    return response.metrics || {};
  } catch (error) {
    console.error("获取域名监控指标失败:", error);
    return {};
  }
};

// 获取路由列表
export const getRoutes = async () => {
  try {
    const response = await api.get("/api/v1/routes");
    return response.routes || [];
  } catch (error) {
    console.error("获取路由列表失败:", error);
    return [];
  }
};

// 创建路由
export const createRoute = async (routeData) => {
  try {
    const response = await api.post("/api/v1/routes", routeData);
    return response;
  } catch (error) {
    console.error("创建路由失败:", error);
    throw error;
  }
};

// 更新路由
export const updateRoute = async (id, routeData) => {
  try {
    const response = await api.put(`/api/v1/routes/${id}`, routeData);
    return response;
  } catch (error) {
    console.error("更新路由失败:", error);
    throw error;
  }
};

// 删除路由
export const deleteRoute = async (id) => {
  try {
    const response = await api.delete(`/api/v1/routes/${id}`);
    return response;
  } catch (error) {
    console.error("删除路由失败:", error);
    throw error;
  }
};

// 获取服务列表
export const getServices = async () => {
  try {
    const response = await api.get("/api/v1/services");
    console.log("服务API原始响应:", response);
    return response;
  } catch (error) {
    console.error("获取服务列表失败:", error);
    return { services: {} };
  }
};

// 获取端点列表
export const getEndpoints = async () => {
  try {
    const response = await api.get("/api/v1/endpoints");
    console.log("端点API原始响应:", response);
    return response;
  } catch (error) {
    console.error("获取端点列表失败:", error);
    return { endpoints: {} };
  }
};

// 健康检查
export const healthCheck = async () => {
  try {
    const response = await api.get("/api/v1/health");
    return response.status === "healthy";
  } catch (error) {
    console.error("健康检查失败:", error);
    return false;
  }
};
