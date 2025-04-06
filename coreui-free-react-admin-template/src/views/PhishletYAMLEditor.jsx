import React, { useEffect, useState } from 'react'
import axios from 'axios'
import {
  CModal,
  CModalHeader,
  CModalTitle,
  CModalBody,
  CModalFooter,
  CButton,
} from '@coreui/react'
import { toast } from 'react-toastify'

import AceEditor from 'react-ace'
import 'ace-builds/src-noconflict/ace'

import ace from 'ace-builds'

// ✅ 直接指向 node_modules 中的资源
ace.config.setModuleUrl('ace/mode/yaml', '/node_modules/ace-builds/src-noconflict/mode-yaml.js')
ace.config.setModuleUrl('ace/theme/github', '/node_modules/ace-builds/src-noconflict/theme-github.js')

const PhishletYAMLEditor = ({ phishletName, visible, onClose, refreshPhishlets }) => {
  const [yamlContent, setYamlContent] = useState('')

  useEffect(() => {
    if (visible && phishletName) {
      axios
        .get(`/api/phishlet/yaml?name=${phishletName}`)
        .then((res) => {
          if (res.data.success) {
            setYamlContent(res.data.content)
          } else {
            toast.error('获取 YAML 失败: ' + res.data.error)
          }
        })
        .catch((err) => {
          toast.error('请求错误: ' + err)
        })
    }
  }, [visible, phishletName])

  const handleSave = async () => {
    try {
      const res = await axios.put('/api/phishlet/yaml', {
        name: phishletName,
        content: yamlContent,
      })
      if (res.data.success) {
        toast.success(res.data.message || 'YAML 保存成功')
        onClose()
        refreshPhishlets()
      } else {
        toast.error(res.data.error || 'YAML 保存失败')
      }
    } catch (error) {
      toast.error('请求错误: ' + error)
    }
  }

  return (
    <CModal visible={visible} onClose={onClose} size="lg">
      <CModalHeader onClose={onClose}>
        <CModalTitle>编辑 YAML: {phishletName}</CModalTitle>
      </CModalHeader>
      <CModalBody>
        <AceEditor
          mode="yaml"
          theme="github"
          width="100%"
          height="400px"
          value={yamlContent}
          onChange={setYamlContent}
          name="phishletYamlEditor"
          editorProps={{ $blockScrolling: true }}
        />
      </CModalBody>
      <CModalFooter>
        <CButton color="secondary" onClick={onClose}>
          取消
        </CButton>
        <CButton color="primary" onClick={handleSave}>
          保存
        </CButton>
      </CModalFooter>
    </CModal>
  )
}

export default PhishletYAMLEditor
