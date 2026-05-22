<template>
  <div>
    <!-- Header -->
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-base font-semibold text-gray-700">批次列表</h3>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">新建批次</el-button>
    </div>

    <!-- Table -->
    <el-card shadow="never" class="border border-gray-200">
      <el-table :data="batches" v-loading="loading" stripe class="w-full">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="note" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'one_time' ? 'warning' : 'success'" size="small">
              {{ row.type === 'one_time' ? '一次性' : '多次使用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="count" label="生成数量" width="90" align="center" />
        <el-table-column label="到期时间" width="160">
          <template #default="{ row }">{{ formatDate(row.expires_at) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button
              type="primary"
              text
              size="small"
              @click="viewKeys(row.id)"
            >
              查看密钥
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      title="新建批次"
      width="480px"
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
        label-position="right"
      >
        <el-form-item label="批次备注" prop="note">
          <el-input v-model="createForm.note" placeholder="输入备注（必填）" />
        </el-form-item>

        <el-form-item label="类型" prop="type">
          <el-select v-model="createForm.type" class="w-full">
            <el-option label="一次性" value="one_time" />
            <el-option label="多次使用" value="multi_use" />
          </el-select>
        </el-form-item>

        <el-form-item label="生成数量" prop="count">
          <el-input-number
            v-model="createForm.count"
            :min="1"
            :max="500"
            class="w-full"
          />
        </el-form-item>

        <el-form-item label="有效期" prop="expires_at">
          <el-date-picker
            v-model="createForm.expires_at"
            type="date"
            placeholder="选择到期日期"
            :disabled-date="disabledDate"
            value-format="YYYY-MM-DD"
            class="w-full"
          />
        </el-form-item>

        <el-form-item
          v-if="createForm.type === 'multi_use'"
          label="最大使用次数"
          prop="max_uses"
        >
          <el-input-number
            v-model="createForm.max_uses"
            :min="2"
            :max="9999"
            class="w-full"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCreate">确认创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getBatches, createBatch } from '../api'
import type { KeyBatch } from '../types'

const router = useRouter()
const batches = ref<KeyBatch[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const submitting = ref(false)
const createFormRef = ref<FormInstance>()

const createForm = reactive({
  note: '',
  type: 'one_time' as 'one_time' | 'multi_use',
  count: 10,
  expires_at: '',
  max_uses: 5,
})

const createRules: FormRules = {
  note: [{ required: true, message: '请输入批次备注', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  count: [{ required: true, message: '请输入生成数量', trigger: 'blur' }],
  expires_at: [{ required: true, message: '请选择有效期', trigger: 'change' }],
}

const disabledDate = (date: Date) => {
  const today = dayjs()
  const minDate = today.add(1, 'day').startOf('day')
  const maxDate = today.add(180, 'day').endOf('day')
  return dayjs(date).isBefore(minDate) || dayjs(date).isAfter(maxDate)
}

const formatDate = (str: string) => {
  if (!str) return '-'
  return dayjs(str).format('YYYY-MM-DD HH:mm')
}

const loadBatches = async () => {
  loading.value = true
  try {
    batches.value = await getBatches()
  } catch {
    ElMessage.error('加载批次失败')
  } finally {
    loading.value = false
  }
}

const viewKeys = (batchId: number) => {
  router.push({ path: '/keys', query: { batch_id: String(batchId) } })
}

const resetForm = () => {
  createFormRef.value?.resetFields()
  createForm.note = ''
  createForm.type = 'one_time'
  createForm.count = 10
  createForm.expires_at = ''
  createForm.max_uses = 5
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload: any = {
      note: createForm.note,
      type: createForm.type,
      count: createForm.count,
      expires_at: createForm.expires_at,
    }
    if (createForm.type === 'multi_use') {
      payload.max_uses = createForm.max_uses
    }
    await createBatch(payload)
    ElMessage.success('批次创建成功')
    showCreateDialog.value = false
    await loadBatches()
  } catch (err: any) {
    const msg = err.response?.data?.message || err.response?.data?.error || '创建失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

onMounted(loadBatches)
</script>
