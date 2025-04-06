import React from 'react'
import CIcon from '@coreui/icons-react'
import {
  cilBell,
  cilCalculator,
  cilChartPie,
  cilCursor,
  cilDescription,
  cilDrop,
  cilExternalLink,
  cilNotes,
  cilPencil,
  cilPuzzle,
  cilSpeedometer,
  cilStar,
  cilBug,
  cilPeople,
  cilList,
  cilSettings,
} from '@coreui/icons'
import { CNavGroup, CNavItem, CNavTitle } from '@coreui/react'

const _nav = [

  {
    component: CNavItem,
    name: 'Dashboard',
    to: '/dashboard',
    icon: <CIcon icon={cilSpeedometer} customClassName="nav-icon" />,
    badge: {
      color: 'info',
      text: 'NEW',
    },
  },

  {
    component: CNavItem,
    name: 'Configuration',
    to: '/config',
    icon: <CIcon icon={cilSettings} customClassName="nav-icon" />,
  },

  {
    component: CNavItem,
    name: 'Projects',
    to: '/projects',
    icon: <CIcon icon={cilList} customClassName="nav-icon" />,
  },

  {
    component: CNavItem,
    name: 'Phishlets',
    to: '/phishlets',
    icon: <CIcon icon={cilBug} customClassName="nav-icon" />,
  },
  
  {
    component: CNavItem,
    name: 'Lures 管理',
    to: '/lures',
    icon: <CIcon icon={cilCursor} customClassName="nav-icon" />,
  },  
  
  {
    component: CNavItem,
    name: 'Sessions',
    to: '/sessions',
    icon: <CIcon icon={cilPeople} customClassName="nav-icon" />,
  },



]

export default _nav
