// src/components/BackgroundAnimation.js
import React, { useRef, useEffect } from 'react'
import * as THREE from 'three'
import { Noise } from 'noisejs'
import './BackgroundAnimation.css'

const BackgroundAnimation = () => {
  const canvasRef = useRef(null)
  const noise = new Noise(Math.random())

  useEffect(() => {
    const canvas = canvasRef.current
    const renderer = new THREE.WebGLRenderer({
      canvas: canvas,
      antialias: true,
    })
    renderer.setPixelRatio(window.devicePixelRatio > 1 ? 2 : 1)
    renderer.setSize(window.innerWidth, window.innerHeight)

    const scene = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(
      45,
      window.innerWidth / window.innerHeight,
      1,
      1000
    )
    camera.position.z = 60

    const length = 30
    let offset = 0
    let prevA = 0
    let isMouseDown = false
    const mouseJump = { x: 0, y: 0 }

    // 修改后的 Spline 使用 BufferGeometry 替代 THREE.Geometry
    function Spline() {
      this.color = Math.floor(Math.random() * 80 + 180)
      const numPoints = 180
      const vertices = []
      const colors = []
      for (let j = 0; j < numPoints; j++) {
        // 初始 x 坐标
        const x = j / numPoints * length * 2 - length
        vertices.push(x, 0, 0)
        const col = new THREE.Color(`hsl(${j * 0.6 + this.color}, 70%, 70%)`)
        colors.push(col.r, col.g, col.b)
      }
      this.geometry = new THREE.BufferGeometry()
      this.geometry.setAttribute('position', new THREE.BufferAttribute(new Float32Array(vertices), 3))
      this.geometry.setAttribute('color', new THREE.BufferAttribute(new Float32Array(colors), 3))
      this.material = new THREE.LineBasicMaterial({ vertexColors: true })
      this.mesh = new THREE.Line(this.geometry, this.material)
      this.speed = (Math.random() + 0.1) * 0.0002
      scene.add(this.mesh)
      // 保存初始 x 坐标数组，便于后续更新
      this.xCoords = []
      for (let j = 0; j < numPoints; j++) {
        const x = j / numPoints * length * 2 - length
        this.xCoords.push(x)
      }
    }

    const splines = []
    for (let i = 0; i < 12; i++) {
      splines.push(new Spline())
    }

    function onResize() {
      camera.aspect = window.innerWidth / window.innerHeight
      camera.updateProjectionMatrix()
      renderer.setSize(window.innerWidth, window.innerHeight)
    }

    function updateColor(a) {
      for (let i = 0; i < splines.length; i++) {
        const spline = splines[i]
        const posAttr = spline.geometry.attributes.position
        const colAttr = spline.geometry.attributes.color
        const numPoints = posAttr.count
        const newColors = new Float32Array(numPoints * 3)
        const colorShift = Math.abs((spline.color - offset * 10) % 360)
        for (let j = 0; j < numPoints; j++) {
          const col = new THREE.Color(`hsl(${j * 0.6 + colorShift},70%,70%)`)
          newColors[j * 3] = col.r
          newColors[j * 3 + 1] = col.g
          newColors[j * 3 + 2] = col.b
        }
        colAttr.array = newColors
        colAttr.needsUpdate = true
      }
    }

    function render(a) {
      requestAnimationFrame(render)
      for (let i = 0; i < splines.length; i++) {
        const spline = splines[i]
        const posAttr = spline.geometry.attributes.position
        const numPoints = posAttr.count
        for (let j = 0; j < numPoints; j++) {
          const x = spline.xCoords[j]
          let y = noise.simplex2(j * 0.05 + i - offset, a * spline.speed) * 8
          let z = noise.simplex2(x * 0.05 + i, a * spline.speed) * 8
          y *= 1 - Math.abs(x / length)
          z *= 1 - Math.abs(x / length)
          posAttr.array[j * 3 + 0] = x
          posAttr.array[j * 3 + 1] = y
          posAttr.array[j * 3 + 2] = z
        }
        posAttr.needsUpdate = true
      }
      scene.rotation.x = a * 0.0003
      if (isMouseDown) {
        mouseJump.x += 0.001
        if (a - prevA > 100) {
          updateColor(a)
          prevA = a
        }
      } else {
        mouseJump.x -= 0.001
      }
      mouseJump.x = Math.max(0, Math.min(0.07, mouseJump.x))
      offset += mouseJump.x
      renderer.render(scene, camera)
    }

    window.addEventListener('resize', onResize)
    window.addEventListener('mousedown', () => { isMouseDown = true })
    window.addEventListener('mouseup', () => { isMouseDown = false })

    requestAnimationFrame(render)

    return () => {
      window.removeEventListener('resize', onResize)
      window.removeEventListener('mousedown', () => { isMouseDown = true })
      window.removeEventListener('mouseup', () => { isMouseDown = false })
      renderer.dispose()
    }
  }, [noise])

  return (
    <div className="background-animation">
      <canvas ref={canvasRef}></canvas>
      <h2>Press down</h2>
    </div>
  )
}

export default BackgroundAnimation
