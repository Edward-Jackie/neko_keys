export interface KeyBatch {
  id: number
  note: string
  count: number
  type: 'one_time' | 'multi_use'
  expires_at: string
  max_uses: number
  created_at: string
  active_count?: number
  total_count?: number
}

export interface Key {
  id: number
  batch_id: number
  key: string
  type: 'one_time' | 'multi_use'
  status: 'inactive' | 'active' | 'expired' | 'disabled'
  expires_at: string
  max_uses: number
  use_count: number
  note: string
  activated_at: string | null
  activated_ip: string
  created_at: string
  batch?: KeyBatch
}

export interface KeyLog {
  id: number
  key_id: number
  ip: string
  action: 'activate' | 'validate'
  success: boolean
  created_at: string
}

export interface DashboardStats {
  total: number
  active: number
  expired: number
  disabled: number
  inactive: number
  today_activations: number
  today_validations: number
  alert_count: number
}

export interface AlertKey {
  key_id: number
  key: string
  ip_count: number
  ips: string[]
}

export interface CreateBatchRequest {
  note: string
  type: 'one_time' | 'multi_use'
  count: number
  expires_at: string
  max_uses?: number
}
