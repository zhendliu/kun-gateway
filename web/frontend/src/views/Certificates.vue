<template>
  <div class="certificates-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>HTTPS证书管理</span>
          <el-button type="primary" @click="showAddDialog = true">
            添加证书
          </el-button>
        </div>
      </template>

      <el-table :data="certificatesList" v-loading="loading" style="width: 100%">
        <el-table-column prop="domain" label="域名" width="200" />
        <el-table-column prop="cert_file" label="证书文件" />
        <el-table-column prop="key_file" label="私钥文件" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button
              type="danger"
              size="small"
              @click="removeCertificate(scope.row.domain)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加证书对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加HTTPS证书"
      width="600px"
    >
      <el-form :model="certificateForm" :rules="rules" ref="certificateFormRef" label-width="100px">
        <el-form-item label="域名" prop="domain">
          <el-input v-model="certificateForm.domain" placeholder="例如: example.com" />
        </el-form-item>
        <el-form-item label="证书文件" prop="certFile">
          <el-upload
            ref="certUploadRef"
            :auto-upload="false"
            :on-change="handleCertChange"
            :before-upload="beforeCertUpload"
            accept=".crt,.pem,.cer"
            :limit="1"
            :file-list="certFileList"
          >
            <el-button type="primary">选择证书文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .crt, .pem, .cer 格式的证书文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="私钥文件" prop="keyFile">
          <el-upload
            ref="keyUploadRef"
            :auto-upload="false"
            :on-change="handleKeyChange"
            :before-upload="beforeKeyUpload"
            accept=".key,.pem"
            :limit="1"
            :file-list="keyFileList"
          >
            <el-button type="primary">选择私钥文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .key, .pem 格式的私钥文件
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="addCertificate" :loading="adding">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCertificates, createCertificate, deleteCertificate } from '@/api/certificates'

const loading = ref(false)
const adding = ref(false)
const showAddDialog = ref(false)
const certificatesList = ref([])

const certificateForm = ref({
  domain: '',
  certFile: null,
  keyFile: null
})

const certFileList = ref([])
const keyFileList = ref([])
const certUploadRef = ref()
const keyUploadRef = ref()

const rules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' }
  ],
  certFile: [
    { required: true, message: '请选择证书文件', trigger: 'change' }
  ],
  keyFile: [
    { required: true, message: '请选择私钥文件', trigger: 'change' }
  ]
}

const certificateFormRef = ref()

const loadCertificates = async () => {
  loading.value = true
  try {
    const response = await getCertificates()
    if (response && response.certificates) {
      certificatesList.value = response.certificates.map(domain => ({
        domain,
        cert_file: `/etc/ssl/certs/${domain}.crt`,
        key_file: `/etc/ssl/certs/${domain}.key`,
        created_at: new Date()
      }))
    } else {
      certificatesList.value = []
    }
  } catch (error) {
    console.error('加载证书失败:', error)
    ElMessage.error('加载证书失败')
    certificatesList.value = []
  } finally {
    loading.value = false
  }
}

const beforeCertUpload = (file) => {
  const isValidType = file.type === 'application/x-x509-ca-cert' || 
                     file.name.endsWith('.crt') || 
                     file.name.endsWith('.pem') || 
                     file.name.endsWith('.cer')
  if (!isValidType) {
    ElMessage.error('证书文件格式不正确，请选择 .crt, .pem, .cer 格式的文件')
    return false
  }
  return false // 阻止自动上传
}

const beforeKeyUpload = (file) => {
  const isValidType = file.type === 'application/x-pem-file' || 
                     file.name.endsWith('.key') || 
                     file.name.endsWith('.pem')
  if (!isValidType) {
    ElMessage.error('私钥文件格式不正确，请选择 .key, .pem 格式的文件')
    return false
  }
  return false // 阻止自动上传
}

const handleCertChange = (file) => {
  certificateForm.value.certFile = file.raw
}

const handleKeyChange = (file) => {
  certificateForm.value.keyFile = file.raw
}

const addCertificate = async () => {
  if (!certificateFormRef.value) return
  
  await certificateFormRef.value.validate(async (valid) => {
    if (valid) {
      if (!certificateForm.value.certFile || !certificateForm.value.keyFile) {
        ElMessage.error('请选择证书文件和私钥文件')
        return
      }

      adding.value = true
      try {
        // 创建FormData对象上传文件
        const formData = new FormData()
        formData.append('domain', certificateForm.value.domain)
        formData.append('cert_file', certificateForm.value.certFile)
        formData.append('key_file', certificateForm.value.keyFile)

        await createCertificate(formData)
        ElMessage.success('证书添加成功')
        showAddDialog.value = false
        
        // 重置表单
        certificateForm.value = { domain: '', certFile: null, keyFile: null }
        certFileList.value = []
        keyFileList.value = []
        if (certUploadRef.value) certUploadRef.value.clearFiles()
        if (keyUploadRef.value) keyUploadRef.value.clearFiles()
        
        loadCertificates()
      } catch (error) {
        console.error('添加证书失败:', error)
        ElMessage.error('添加证书失败')
      } finally {
        adding.value = false
      }
    }
  })
}

const removeCertificate = async (domain) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除域名 ${domain} 的证书吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await deleteCertificate(domain)
    ElMessage.success('证书删除成功')
    loadCertificates()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除证书失败:', error)
      ElMessage.error('删除证书失败')
    }
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  loadCertificates()
})
</script>

<style scoped>
.certificates-page {
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

.el-upload__tip {
  color: #909399;
  font-size: 12px;
  margin-top: 5px;
}
</style> 