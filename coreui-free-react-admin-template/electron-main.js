// electron-main.js
const { app, BrowserWindow } = require('electron')
const path = require('path')

function createWindow() {
  // 创建浏览器窗口
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false, // 是否启用 nodeIntegration
      contextIsolation: true, // 建议开启隔离
    },
  })

  // 加载打包后前端的 index.html
  // 假设打包输出文件夹是 "build"
  const indexPath = path.join(__dirname, 'build', 'index.html')
  win.loadURL(`file://${indexPath}`)

  // 打开调试工具（生产环境可以注释掉）
  // win.webContents.openDevTools()
}

// 当 Electron 完成初始化并准备创建浏览器窗口时调用
app.whenReady().then(() => {
  createWindow()

  app.on('activate', () => {
    // macOS 上，当点击dock图标并且没有其它打开的窗口时，重新创建一个窗口
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

// 所有窗口关闭时退出应用 (在 macOS 上除外)
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
