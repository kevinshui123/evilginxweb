#!/bin/bash

# ==== CONFIG区域 ====
DOMAIN="admin.example.com"             # 改成你自己的域名
BACKEND_PORT="8080"                    # Evilginx后端监听端口
EMAIL="you@example.com"               # 用于注册 Let's Encrypt

# 自动识别 build 目录路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(realpath "$SCRIPT_DIR/build")"

# ==== 1. 安装依赖 ====
sudo apt update
sudo apt install nginx certbot python3-certbot-nginx -y

# ==== 2. 创建前端目录（如果不存在） ====
sudo mkdir -p "$FRONTEND_DIR"
sudo chown -R $USER:$USER "$FRONTEND_DIR"

# ==== 3. 配置 Nginx ====
NGINX_CONF="/etc/nginx/sites-available/$DOMAIN"

sudo tee "$NGINX_CONF" > /dev/null <<EOF
server {
  listen 80;
  server_name $DOMAIN;

  root $FRONTEND_DIR;
  index index.html;

  location / {
    try_files \$uri /index.html;
  }

  location /api/ {
    proxy_pass http://127.0.0.1:$BACKEND_PORT;
    proxy_http_version 1.1;
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
  }
}
EOF

# 启用站点配置
sudo ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/"
sudo nginx -t && sudo systemctl reload nginx

# ==== 4. 自动申请 SSL 证书 ====
sudo certbot --nginx --non-interactive --agree-tos -m "$EMAIL" -d "$DOMAIN"

# ==== 5. 设置自动续期（推荐） ====
echo "0 3 * * * root certbot renew --quiet" | sudo tee /etc/cron.d/certbot-auto-renew

echo ""
echo "✅ 部署完成！现在你可以访问: https://$DOMAIN"
