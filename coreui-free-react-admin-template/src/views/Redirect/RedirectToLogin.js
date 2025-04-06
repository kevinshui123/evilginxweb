import React from 'react'
import { Navigate } from 'react-router-dom'

const RedirectToLogin = () => {
  return <Navigate to="/login" replace />
}

export default RedirectToLogin
