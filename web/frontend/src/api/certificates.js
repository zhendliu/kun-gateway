import request from '@/utils/request'

// 获取证书列表
export function getCertificates() {
  return request({
    url: '/api/v1/certificates',
    method: 'get'
  })
}

// 创建证书（支持文件上传）
export function createCertificate(data) {
  return request({
    url: '/api/v1/certificates',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 删除证书
export function deleteCertificate(domain) {
  return request({
    url: `/api/v1/certificates/${domain}`,
    method: 'delete'
  })
} 