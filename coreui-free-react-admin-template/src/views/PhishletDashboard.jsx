import React, { useEffect, useState } from 'react'
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
  CFormSwitch,
  CModal,
  CModalHeader,
  CModalTitle,
  CModalBody,
  CModalFooter,
  CFormInput,
  CFormCheck,
  CSpinner,
} from '@coreui/react'
import { toast } from 'react-toastify'
import PhishletYAMLEditor from './PhishletYAMLEditor'

const PhishletDashboard = () => {
  const [phishlets, setPhishlets] = useState([])
  const [loading, setLoading] = useState(true)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [currentPhishlet, setCurrentPhishlet] = useState({
    name: '',
    hostname: '',
    visibility: true,
    unauth_url: '',
  })
  const [yamlModalVisible, setYamlModalVisible] = useState(false)
  const [currentYAMLPhishlet, setCurrentYAMLPhishlet] = useState('')
  const [uploadModalVisible, setUploadModalVisible] = useState(false)
  const [uploadFile, setUploadFile] = useState(null)

  useEffect(() => {
    fetchPhishlets()
  }, [])

  const fetchPhishlets = async () => {
    setLoading(true)
    try {
      const res = await axios.get('/api/phishlet')
      // 确保返回值为数组，否则置为空数组
      setPhishlets(Array.isArray(res.data) ? res.data : [])
    } catch (error) {
      console.error('Failed to fetch phishlets:', error)
      setPhishlets([])
    } finally {
      setLoading(false)
    }
  }

  const handleDisable = async (name) => {
    try {
      await axios.post('/api/phishlet/disable', { name })
      fetchPhishlets()
    } catch (error) {
      console.error('Failed to disable phishlet:', error)
    }
  }

  const handleEnable = async (name) => {
    try {
      const { data } = await axios.post('/api/phishlet/enable', { name })
      console.log('enable response:', data)
      if (!data.success) {
        toast.error('Enable failed: ' + data.error)
      }
      fetchPhishlets()
    } catch (error) {
      console.error('Failed to enable phishlet:', error)
      toast.error('Failed to enable phishlet: ' + error)
    }
  }

  const handleToggle = (p) => {
    if (p.enabled) {
      handleDisable(p.name)
    } else {
      handleEnable(p.name)
    }
  }

  const handleEditClick = (p) => {
    setCurrentPhishlet({
      name: p.name,
      hostname: p.hostname,
      visibility: !p.hidden, // 若 p.hidden 为 true，则 visibility 为 false
      unauth_url: p.unauth_url || '',
    })
    setEditModalVisible(true)
  }

  const handleSaveEdit = async () => {
    try {
      const { data } = await axios.put('/api/phishlet/edit', currentPhishlet)
      if (data.success) {
        toast.success(data.message || '保存成功！')
      } else {
        toast.error(data.error || '保存失败')
      }
      setEditModalVisible(false)
      fetchPhishlets()
    } catch (error) {
      toast.error('请求错误：' + error)
    }
  }

  const handleDeletePhishlet = async (name) => {
    if (!window.confirm(`确定要删除 phishlet: ${name} 吗？`)) return

    try {
      const res = await axios.delete(`/api/phishlet/delete?name=${name}`)
      if (res.data.success) {
        toast.success('删除成功')
        fetchPhishlets()
      } else {
        toast.error('删除失败：' + res.data.error)
      }
    } catch (error) {
      toast.error('请求错误：' + error.message)
    }
  }

  const handleEditYAMLClick = (p) => {
    setCurrentYAMLPhishlet(p.name)
    setYamlModalVisible(true)
  }

  const handleUpload = async () => {
    if (!uploadFile) return
    const formData = new FormData()
    formData.append('phishlet_file', uploadFile)

    try {
      const res = await axios.post('/api/phishlet/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      if (res.data.success) {
        toast.success('上传成功')
        setUploadModalVisible(false)
        setUploadFile(null)
        fetchPhishlets()
      } else {
        toast.error('上传失败：' + res.data.error)
      }
    } catch (error) {
      console.error('Failed to upload phishlet:', error)
      toast.error('上传出错：' + error.message)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <CCard>
        <CCardHeader>
          <div className="d-flex justify-content-between">
            <span>Phishlets Dashboard</span>
            <CButton color="primary" onClick={() => setUploadModalVisible(true)}>上传 YAML</CButton>
          </div>
        </CCardHeader>
        <CCardBody>
          {loading ? (
            <div className="text-center py-4">
              <CSpinner color="primary" />
            </div>
          ) : phishlets.length === 0 ? (
            <div className="text-center text-muted p-4">
              暂无可用的 Phishlet，请先上传或创建。
            </div>
          ) : (
            <CTable>
              <CTableHead>
                <CTableRow>
                  <CTableHeaderCell>Name</CTableHeaderCell>
                  <CTableHeaderCell>Enabled</CTableHeaderCell>
                  <CTableHeaderCell>Hostname</CTableHeaderCell>
                  <CTableHeaderCell>操作</CTableHeaderCell>
                </CTableRow>
              </CTableHead>
              <CTableBody>
                {phishlets.map((p) => (
                  <CTableRow key={p.name}>
                    <CTableDataCell>{p.name}</CTableDataCell>
                    <CTableDataCell>
                      <CFormSwitch
                        checked={p.enabled}
                        onChange={() => {
                          console.log('switch clicked for', p.name)
                          handleToggle(p)
                        }}
                      />
                    </CTableDataCell>
                    <CTableDataCell>{p.hostname}</CTableDataCell>
                    <CTableDataCell>
                      <CButton color="warning" size="sm" onClick={() => handleEditClick(p)}>
                        编辑
                      </CButton>{' '}
                      <CButton color="info" size="sm" onClick={() => handleEditYAMLClick(p)}>
                        编辑 YAML
                      </CButton>{' '}
                      <CButton color="danger" size="sm" onClick={() => handleDeletePhishlet(p.name)}>
                        删除
                      </CButton>
                    </CTableDataCell>
                  </CTableRow>
                ))}
              </CTableBody>
            </CTable>
          )}
        </CCardBody>
      </CCard>

      {/* 编辑 Phishlet Modal */}
      <CModal visible={editModalVisible} onClose={() => setEditModalVisible(false)}>
        <CModalHeader onClose={() => setEditModalVisible(false)}>
          <CModalTitle>编辑 Phishlet</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <div className="mb-3">
            <label>Hostname</label>
            <CFormInput
              type="text"
              value={currentPhishlet.hostname}
              onChange={(e) =>
                setCurrentPhishlet({ ...currentPhishlet, hostname: e.target.value })
              }
            />
          </div>
          <div className="mb-3">
            <label>Visibility</label>
            <CFormSwitch
              label="Visible"
              checked={currentPhishlet.visibility}
              onChange={(e) =>
                setCurrentPhishlet({ ...currentPhishlet, visibility: e.target.checked })
              }
            />
          </div>
          <div className="mb-3">
            <label>Unauth URL</label>
            <CFormInput
              type="text"
              value={currentPhishlet.unauth_url}
              onChange={(e) =>
                setCurrentPhishlet({ ...currentPhishlet, unauth_url: e.target.value })
              }
            />
          </div>
        </CModalBody>
        <CModalFooter>
          <CButton color="secondary" onClick={() => setEditModalVisible(false)}>
            取消
          </CButton>
          <CButton color="primary" onClick={handleSaveEdit}>
            保存
          </CButton>
        </CModalFooter>
      </CModal>

      {/* YAML 编辑 Modal */}
      {yamlModalVisible && (
        <PhishletYAMLEditor
          phishletName={currentYAMLPhishlet}
          visible={yamlModalVisible}
          onClose={() => setYamlModalVisible(false)}
          refreshPhishlets={fetchPhishlets}
        />
      )}

      {/* 上传 YAML Modal */}
      <CModal visible={uploadModalVisible} onClose={() => setUploadModalVisible(false)}>
        <CModalHeader onClose={() => setUploadModalVisible(false)}>
          <CModalTitle>上传 Phishlet YAML</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <CFormInput
            type="file"
            onChange={(e) => setUploadFile(e.target.files[0])}
          />
        </CModalBody>
        <CModalFooter>
          <CButton color="secondary" onClick={() => setUploadModalVisible(false)}>
            取消
          </CButton>
          <CButton color="primary" onClick={handleUpload}>
            上传
          </CButton>
        </CModalFooter>
      </CModal>
    </div>
  )
}

export default PhishletDashboard
