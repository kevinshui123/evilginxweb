// src/views/Login/Login.js
import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  CCard,
  CCardBody,
  CCardHeader,
  CForm,
  CFormInput,
  CButton,
  CFormLabel,
  CContainer,
  CRow,
  CCol,
} from '@coreui/react'
import './Login.css'
import BackgroundAnimation from '../../../components/BackgroundAnimation'

const Login = () => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const navigate = useNavigate()

  const handleLogin = (e) => {
    e.preventDefault()
    if (username === 'admin' && password === 'password') {
      localStorage.setItem('isLoggedIn', 'true')
      navigate('/dashboard')
    } else {
      alert('用户名或密码错误')
    }
  }

  return (
    <div className="login-page">
      {/* 背景动画组件 */}
      <BackgroundAnimation />
      {/* 登录表单部分，保证层级高于背景动画 */}
      <CContainer
        className="h-100 d-flex justify-content-center align-items-center"
        style={{ position: 'relative', zIndex: 1 }}
      >
        <CRow className="w-100 justify-content-center">
          <CCol md={6} lg={4}>
            <CCard>
              <CCardHeader>登录</CCardHeader>
              <CCardBody>
                <CForm onSubmit={handleLogin}>
                  <div className="mb-3">
                    <CFormLabel htmlFor="username">用户名</CFormLabel>
                    <CFormInput
                      id="username"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      required
                    />
                  </div>
                  <div className="mb-3">
                    <CFormLabel htmlFor="password">密码</CFormLabel>
                    <CFormInput
                      type="password"
                      id="password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      required
                    />
                  </div>
                  <CButton color="primary" type="submit" className="w-100">
                    登录
                  </CButton>
                </CForm>
              </CCardBody>
            </CCard>
          </CCol>
        </CRow>
      </CContainer>
    </div>
  )
}

export default Login
