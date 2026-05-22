import axios from 'axios'
import type { AxiosInstance } from 'axios'
import type { KeyBatch, Key, KeyLog, DashboardStats, AlertKey, CreateBatchRequest } from '../types'

const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:9002'

const request: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000,
})

// Request interceptor: attach JWT token
request.interceptors.request.use((config) => {
  // Lazy import to avoid circular dependency
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: handle 401
request.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth
export function login(username: string, password: string): Promise<{ token: string; username: string }> {
  return request.post('/api/admin/login', { username, password }).then((r) => r.data)
}

// Dashboard
export function getDashboard(): Promise<DashboardStats> {
  return request.get('/api/admin/dashboard').then((r) => r.data)
}

// Batches
export function getBatches(): Promise<KeyBatch[]> {
  return request.get('/api/admin/batches').then((r) => r.data)
}

export function createBatch(data: CreateBatchRequest): Promise<KeyBatch> {
  return request.post('/api/admin/batches', data).then((r) => r.data)
}

// Keys
export function getKeys(params: {
  batch_id?: number
  status?: string
  keyword?: string
  page?: number
  page_size?: number
}): Promise<{ list: Key[]; total: number }> {
  return request.get('/api/admin/keys', { params }).then((r) => r.data)
}

export function getKey(id: number): Promise<Key & { logs: KeyLog[] }> {
  return request.get(`/api/admin/keys/${id}`).then((r) => r.data)
}

export function updateKey(id: number, data: { note?: string; status?: string }): Promise<Key> {
  return request.put(`/api/admin/keys/${id}`, data).then((r) => r.data)
}

export function getKeyLogs(id: number): Promise<KeyLog[]> {
  return request.get(`/api/admin/keys/${id}/logs`).then((r) => r.data)
}

// Alerts
export function getAlerts(): Promise<AlertKey[]> {
  return request.get('/api/admin/alerts').then((r) => r.data)
}
