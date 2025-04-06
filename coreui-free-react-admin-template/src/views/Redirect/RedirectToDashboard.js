import React from 'react'
import { Navigate } from 'react-router-dom'

const RedirectToDashboard = () => {
  return <Navigate to="/dashboard" replace />
}

export default RedirectToDashboard
