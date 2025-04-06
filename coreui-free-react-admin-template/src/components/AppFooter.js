import React from 'react'
import { CFooter } from '@coreui/react'

const AppFooter = () => {
  return (
    <CFooter className="px-4">
      <div className="mx-auto">
        <span>&copy; 2025 松子壳工具（SquirrelSnare）管理平台</span>
      </div>
    </CFooter>
  )
}

export default React.memo(AppFooter)
