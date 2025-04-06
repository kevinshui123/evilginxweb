import React, { useEffect, useState } from 'react'
import axios from 'axios'
import {
  CCard,
  CCardBody,
  CCol,
  CRow,
  CWidgetStatsB,
  CFormSwitch,
} from '@coreui/react'

import WidgetsDropdown from '../widgets/WidgetsDropdown'
import MainChart from './MainChart'

const Dashboard = () => {
  const [stats, setStats] = useState({
    active_phishlets: 0,
    total_sessions: 0,
    blacklist_ips: 0,
  })

  // 用来记录 evilginx2 程序的状态
  const [evilginx2Running, setEvilginx2Running] = useState(false)

  // 获取 dashboard 统计数据
  useEffect(() => {
    axios.get('/api/dashboard/stats').then((res) => {
      setStats(res.data)
    })
  }, [])

  // 获取 evilginx2 当前状态
  useEffect(() => {
    axios.get('/api/evilginx2/status')
      .then((res) => {
        // 假设返回 { running: true/false }
        setEvilginx2Running(res.data.running)
      })
      .catch((err) => {
        console.error('获取 evilginx2 状态失败:', err)
      })
  }, [])

  // 切换 evilginx2 状态
  const toggleEvilginx2 = () => {
    // 根据当前状态决定调用启动或关闭接口
    const endpoint = evilginx2Running ? 'stop' : 'start'
    axios.post(`/api/evilginx2/${endpoint}`)
      .then(() => {
        setEvilginx2Running(!evilginx2Running)
      })
      .catch((err) => {
        console.error(`切换 evilginx2 状态失败: ${endpoint}`, err)
      })
  }

  return (
    <>
      <h3 className="mb-4">Dashboard 总览</h3>
      <CRow className="mb-4">
        <CCol sm={4}>
          <CWidgetStatsB className="mb-3" color="primary" value={stats.active_phishlets} title="活跃 Phishlets" />
        </CCol>
        <CCol sm={4}>
          <CWidgetStatsB className="mb-3" color="info" value={stats.total_sessions} title="总 Sessions" />
        </CCol>
        <CCol sm={4}>
          <CWidgetStatsB className="mb-3" color="danger" value={stats.blacklist_ips} title="黑名单 IP 数" />
        </CCol>
      </CRow>

      {/* 新增：evilginx2 开关 */}
      <CRow className="mb-4">
        <CCol sm={12}>
          <CFormSwitch
            id="evilginx2-switch"
            checked={evilginx2Running}
            onChange={toggleEvilginx2}
            label={`Evilginx2 程序 ${evilginx2Running ? '运行中' : '已关闭'}`}
          />
        </CCol>
      </CRow>

      <CCard className="mb-4">
        <CCardBody>
          <CRow>
            <CCol sm={5}>
              <h4 id="traffic" className="card-title mb-0">
                会话趋势图
              </h4>
            </CCol>
          </CRow>
          <MainChart />
        </CCardBody>
      </CCard>
    </>
  )
}

export default Dashboard
