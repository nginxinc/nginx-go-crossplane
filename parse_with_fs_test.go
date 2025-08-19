package crossplane

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

const nginx_conf = `# 全局配置段
user  nginx;
worker_processes  auto;  # 自动根据CPU核心数设置worker进程数
error_log  /var/log/nginx/error.log warn;  # 错误日志路径和级别
pid        /var/run/nginx.pid;

# 事件模块配置
events {
    worker_connections  1024;  # 每个worker进程的最大连接数
    multi_accept        on;    # 一次接受所有新连接
    use                 epoll; # 使用epoll高效模型(Linux)
}

# HTTP模块配置
http {
    include       mime.types;  # MIME类型定义
    default_type  application/octet-stream;  # 默认MIME类型

    # 日志格式定义
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    log_format  detailed '$remote_addr - $remote_user [$time_local] "$request" '
                        '$status $body_bytes_sent "$http_referer" '
                        '"$http_user_agent" "$http_x_forwarded_for" '
                        'rt=$request_time uct="$upstream_connect_time" uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log  /var/log/nginx/access.log  main;  # 访问日志

    sendfile        on;  # 高效文件传输
    tcp_nopush     on;  # 优化数据包发送
    tcp_nodelay    on;  # 禁用Nagle算法

    keepalive_timeout  65;  # 保持连接超时时间
    types_hash_max_size 2048;

    # 启用gzip压缩
    gzip  on;
    gzip_disable "msie6";
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_buffers 16 8k;
    gzip_http_version 1.1;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # 安全相关头部
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # 上传文件大小限制
    client_max_body_size 100M;

    # 负载均衡上游服务器配置
    upstream backend {
        least_conn;  # 最少连接负载均衡算法
        server backend1.example.com:8080 weight=5;
        server backend2.example.com:8080;
        server backup.backend.example.com:8080 backup;  # 备份服务器
    }

    # 另一个上游配置 - 用于WebSocket
    upstream websocket {
        server ws1.example.com:8080;
        server ws2.example.com:8080;
    }

    # 静态文件缓存路径配置
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=STATIC:10m inactive=24h max_size=1g use_temp_path=off;

    # 虚拟主机配置 (HTTP)
    server {
        listen       80;
        server_name  example.com www.example.com;
        root         /var/www/html;

        # 全局错误页面
        error_page 404 /404.html;
        error_page 500 502 503 504 /50x.html;

        # 重定向所有HTTP到HTTPS
        return 301 https://$host$request_uri;
    }

    # 虚拟主机配置 (HTTPS)
    server {
        listen       443 ssl http2;
        server_name  example.com www.example.com;

        # SSL证书配置
        ssl_certificate      /etc/ssl/certs/example.com.crt;
        ssl_certificate_key  /etc/ssl/private/example.com.key;
        ssl_trusted_certificate /etc/ssl/certs/example.com.ca.crt;

        # SSL优化配置
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256';
        ssl_prefer_server_ciphers on;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;
        ssl_session_tickets off;
        ssl_stapling on;
        ssl_stapling_verify on;

        # OCSP Stapling
        resolver 8.8.8.8 8.8.4.4 valid=300s;
        resolver_timeout 5s;

        root /var/www/html;

        # 静态文件服务配置
        location / {
            try_files $uri $uri/ /index.html;
            expires 1d;  # 缓存控制
            add_header Cache-Control "public";
        }

        # 静态资源目录
        location /static/ {
            alias /var/www/static/;
            expires 1y;
            access_log off;
            add_header Cache-Control "public";
        }

        # API反向代理配置
        location /api/ {
            proxy_pass http://backend/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache STATIC;
            proxy_cache_valid 200 1h;
            proxy_cache_use_stale error timeout invalid_header updating http_500 http_502 http_503 http_504;
        }

        # WebSocket配置
        location /ws/ {
            proxy_pass http://websocket;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "Upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_read_timeout 86400;  # WebSocket长连接超时
        }

        # 禁止访问隐藏文件
        location ~ /\. {
            deny all;
            access_log off;
            log_not_found off;
        }

        # 基本认证保护的管理区域
        location /admin/ {
            auth_basic "Admin Area";
            auth_basic_user_file /etc/nginx/.htpasswd;
            try_files $uri $uri/ /admin/index.html;
        }

        # 健康检查端点
        location /health {
            access_log off;
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }

        # 禁止常见漏洞扫描
        location ~* (wp-admin|wp-login|\.git) {
            deny all;
        }
    }

    # 子域名配置
    server {
        listen 80;
        server_name blog.example.com;
        
        location / {
            proxy_pass http://blog-backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }

    # 重定向配置
    server {
        listen 80;
        server_name old.example.com;
        return 301 https://example.com$request_uri;
    }

    # 默认服务器配置 - 捕获所有未匹配的请求
    server {
        listen 80 default_server;
        listen [::]:80 default_server;
        server_name _;
        return 444;  # 关闭连接而不发送响应头
    }
}

# 邮件代理配置示例 (可选)
mail {
    server_name mail.example.com;
    auth_http   localhost:9000/auth;
    
    proxy_pass_error_message on;
    
    server {
        listen     25;
        protocol   smtp;
        smtp_auth  login plain cram-md5;
    }
    
    server {
        listen    110;
        protocol  pop3;
        pop3_auth plain apop cram-md5;
    }
    
    server {
        listen    143;
        protocol  imap;
    }
}

# TCP/UDP代理配置示例 (Nginx Plus或1.9.0+)
stream {
    upstream dns_servers {
        server 192.168.1.1:53;
        server 192.168.1.2:53;
    }
    
    server {
        listen 53 udp;
        proxy_pass dns_servers;
        proxy_timeout 1s;
        proxy_responses 1;
    }
    
    server {
        listen 3306;
        proxy_pass db_master;
    }
}`

const mime_types = ``

func Test_ParseWithMemFs(t *testing.T) {
	memfs := fstest.MapFS{
		"nginx.conf": {
			Data: []byte(nginx_conf),
		},
		"mime.types": {
			Data: []byte(mime_types),
		},
	}

	payload, err := ParseWithMemFs(memfs, "nginx.conf", &MemFsParseOptions{})

	if err != nil {
		t.Fatal(err)
	}

	if data, err := json.Marshal(payload); err == nil {
		t.Log(string(data))
	}
}
