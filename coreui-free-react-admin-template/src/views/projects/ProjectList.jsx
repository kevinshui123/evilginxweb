// ProjectList.jsx
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
  CModal,
  CModalHeader,
  CModalTitle,
  CModalBody,
  CModalFooter,
  CFormInput,
  CFormTextarea,
  CRow,
  CCol,
} from '@coreui/react'

const ProjectList = () => {
  // 状态声明
  const [projects, setProjects] = useState([])
  const [modalVisible, setModalVisible] = useState(false)             // 新建项目弹窗
  const [editModalVisible, setEditModalVisible] = useState(false)       // 编辑项目弹窗
  const [currentProject, setCurrentProject] = useState()            // 当前编辑的项目
  const [newProject, setNewProject] = useState({                        // 新项目信息
    name: '',
    description: '',
    phishlets: '',
    lures: '',
    domain: '',
  })

  // 获取项目数据
  const fetchProjects = async () => {
    try {
      const res = await axios.get('/api/projects')
      setProjects(res.data)
    } catch (error) {
      console.error('Failed to fetch projects:', error)
    }
  }

  useEffect(() => {
    fetch('/api/projects')
      .then(res => res.json())
      .then(data => setProjects(data || [])) // 避免 set null
  }, [])
  

  // 点击编辑按钮时，设置当前项目并显示编辑弹窗
  const handleEdit = (project) => {
    setCurrentProject(project)
    setEditModalVisible(true)
  }

  // 新建项目
  const handleCreate = async () => {
    try {
      await axios.post('/api/projects', newProject)
      setModalVisible(false)
      setNewProject({ name: '', description: '', phishlets: '', lures: '', domain: '' })
      fetchProjects()
    } catch (error) {
      console.error('Failed to create project:', error)
    }
  }

  // 删除项目
  const handleDelete = async (id) => {
    try {
      await axios.delete(`/api/projects?id=${id}`)
      fetchProjects()
    } catch (error) {
      console.error('Failed to delete project:', error)
    }
  }

  // 保存编辑后的项目（更新操作）
  const handleSaveEdit = async () => {
    try {
      await axios.put('/api/projects', currentProject)
      setEditModalVisible(false)
      fetchProjects()
    } catch (error) {
      console.error('Failed to update project:', error)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <CCard>
        <CCardHeader>
          项目管理
          <CButton color="primary" className="float-end" onClick={() => setModalVisible(true)}>
            新建项目
          </CButton>
        </CCardHeader>
        <CCardBody>
          <CTable striped responsive>
            <CTableHead>
              <CTableRow>
                <CTableHeaderCell>ID</CTableHeaderCell>
                <CTableHeaderCell>名称</CTableHeaderCell>
                <CTableHeaderCell>描述</CTableHeaderCell>
                <CTableHeaderCell>关联 Phishlets</CTableHeaderCell>
                <CTableHeaderCell>关联 Lures</CTableHeaderCell>
                <CTableHeaderCell>域名</CTableHeaderCell>
                <CTableHeaderCell>操作</CTableHeaderCell>
              </CTableRow>
            </CTableHead>
            <CTableBody>
              {projects.map((project) => (
                <CTableRow key={project.id}>
                  <CTableDataCell>{project.id}</CTableDataCell>
                  <CTableDataCell>{project.name}</CTableDataCell>
                  <CTableDataCell>{project.description}</CTableDataCell>
                  <CTableDataCell>{project.phishlets}</CTableDataCell>
                  <CTableDataCell>{project.lures}</CTableDataCell>
                  <CTableDataCell>{project.domain}</CTableDataCell>
                  <CTableDataCell>
                    <CButton color="warning" size="sm" onClick={() => handleEdit(project)}>
                      编辑
                    </CButton>{' ' }
                    <CButton color="danger" size="sm" onClick={() => handleDelete(project.id)}>
                      删除
                    </CButton>
                  </CTableDataCell>
                </CTableRow>
              ))}
            </CTableBody>
          </CTable>
        </CCardBody>
      </CCard>

      {/* 新建项目弹窗 */}
      <CModal visible={modalVisible} onClose={() => setModalVisible(false)}>
        <CModalHeader onClose={() => setModalVisible(false)}>
          <CModalTitle>新建项目</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <CRow className="mb-3">
            <CCol md={12}>
              <CFormInput
                type="text"
                label="项目名称"
                value={newProject.name}
                onChange={(e) => setNewProject({ ...newProject, name: e.target.value })} />
            </CCol>
          </CRow>
          <CRow className="mb-3">
            <CCol md={12}>
              <CFormTextarea
                label="描述"
                value={newProject.description}
                onChange={(e) => setNewProject({ ...newProject, description: e.target.value })} />
            </CCol>
          </CRow>
          <CRow className="mb-3">
            <CCol md={12}>
              <CFormInput
                type="text"
                label="关联 Phishlets (逗号分隔)"
                value={newProject.phishlets}
                onChange={(e) => setNewProject({ ...newProject, phishlets: e.target.value })} />
            </CCol>
          </CRow>
          <CRow className="mb-3">
            <CCol md={12}>
              <CFormInput
                type="text"
                label="关联 Lures (逗号分隔)"
                value={newProject.lures}
                onChange={(e) => setNewProject({ ...newProject, lures: e.target.value })} />
            </CCol>
          </CRow>
          <CRow className="mb-3">
            <CCol md={12}>
              <CFormInput
                type="text"
                label="域名"
                value={newProject.domain}
                onChange={(e) => setNewProject({ ...newProject, domain: e.target.value })} />
            </CCol>
          </CRow>
        </CModalBody>
        <CModalFooter>
          <CButton color="secondary" onClick={() => setModalVisible(false)}>
            取消
          </CButton>
          <CButton color="primary" onClick={handleCreate}>
            创建
          </CButton>
        </CModalFooter>
      </CModal>

      {/* 编辑项目弹窗 */}
      <CModal visible={editModalVisible} onClose={() => setEditModalVisible(false)}>
        <CModalHeader onClose={() => setEditModalVisible(false)}>
          <CModalTitle>编辑项目</CModalTitle>
        </CModalHeader>
        <CModalBody>
          {currentProject && (
            <>
              <CFormInput
                type="text"
                label="项目名称"
                value={currentProject.name}
                onChange={(e) => setCurrentProject({ ...currentProject, name: e.target.value })} />
              <CFormTextarea
                label="描述"
                value={currentProject.description}
                onChange={(e) => setCurrentProject({ ...currentProject, description: e.target.value })} />
              <CFormInput
                type="text"
                label="关联 Phishlets (逗号分隔)"
                value={currentProject.phishlets}
                onChange={(e) => setCurrentProject({ ...currentProject, phishlets: e.target.value })} />
              <CFormInput
                type="text"
                label="关联 Lures (逗号分隔)"
                value={currentProject.lures}
                onChange={(e) => setCurrentProject({ ...currentProject, lures: e.target.value })} />
              <CFormInput
                type="text"
                label="域名"
                value={currentProject.domain}
                onChange={(e) => setCurrentProject({ ...currentProject, domain: e.target.value })} />
            </>
          )}
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
    </div>
  )
}

export default ProjectList
