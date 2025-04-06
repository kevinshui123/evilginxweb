// views/config/ConfigPanel.js
import React, { useEffect, useState } from 'react'
import axios from 'axios'
import { CCard, CCardBody, CCardHeader, CForm, CFormInput, CFormSwitch, CButton } from '@coreui/react'
import { toast } from 'react-toastify'

const ConfigPanel = () => {
  const [config, setConfig] = useState({
    domain: '',
    external_ip: '',
    bind_ip: '',
    https_port: 443,
    dns_port: 53,
    unauth_url: '',
    autocert: true,
  })

  useEffect(() => {
    fetchConfig()
  }, [])

  const fetchConfig = async () => {
    try {
      const { data } = await axios.get('/api/config')
      setConfig(data)
    } catch (error) {
      toast.error('获取配置失败')
    }
  }

  const handleChange = (key, value) => {
    setConfig({ ...config, [key]: value })
  }

  const handleSave = async () => {
    try {
      const { data } = await axios.put('/api/config', config)
      if (data.success) {
        toast.success(data.message || '保存成功')
      } else {
        toast.error(data.error || '保存失败')
      }
    } catch (error) {
      toast.error('保存请求失败：' + error)
    }
  }

  return (
    <CCard className="mt-4">
      <CCardHeader>全局配置</CCardHeader>
      <CCardBody>
        <CForm className="row g-3">
          <CFormInput label="绑定域名" value={config.domain} onChange={(e) => handleChange('domain', e.target.value)} />
          <CFormInput label="外部IP" value={config.external_ip} onChange={(e) => handleChange('external_ip', e.target.value)} />
          <CFormInput label="绑定IP" value={config.bind_ip} onChange={(e) => handleChange('bind_ip', e.target.value)} />
          <CFormInput label="HTTPS端口" type="number" value={config.https_port} onChange={(e) => handleChange('https_port', Number(e.target.value))} />
          <CFormInput label="DNS端口" type="number" value={config.dns_port} onChange={(e) => handleChange('dns_port', Number(e.target.value))} />
          <CFormInput label="Unauth URL" value={config.unauth_url} onChange={(e) => handleChange('unauth_url', e.target.value)} />
          <div className="form-check form-switch">
            <input
              className="form-check-input"
              type="checkbox"
              checked={config.autocert}
              onChange={(e) => handleChange('autocert', e.target.checked)}
              id="autocertSwitch"
            />
            <label className="form-check-label" htmlFor="autocertSwitch">启用自动证书 (autocert)</label>
          </div>
          <div>
            <CButton color="primary" onClick={handleSave}>保存配置</CButton>
          </div>
        </CForm>
      </CCardBody>
    </CCard>
  )
}

export default ConfigPanel
