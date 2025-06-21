
# nginx

```bash
nano /etc/nginx/conf.d/slcw.conf
sudo nginx -t
systemctl restart nginx.service
```

# ssh туннель 
```bash
ssh -R 0.0.0.0:10081:localhost:8081 root@209.145.52.169 -N -p 2222 -o ServerAliveInterval=10 -o ServerAliveCountMax=5
```
