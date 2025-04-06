import axios from 'axios'

const API_BASE = '/api' // Evilginx Admin 地址

export const fetchPhishlets = async () => {
  const response = await axios.get(`${API_BASE}/api/phishlets`)
  return response.data
}

export const disablePhishlet = async (name) => {
  const response = await axios.post(`${API_BASE}/api/phishlets/disable`, { name })
  return response.data
}

export const fetchDashboardStats = async () => {
  const response = await axios.get(`${API_BASE}/api/dashboard/stats`)
  return response.data
}

export const fetchSessions = async () => {
  const response = await axios.get(`${API_BASE}/api/sessions`)
  return response.data
}
