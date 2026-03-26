#!/bin/bash

# สร้างไฟล์ที่จำเป็นถ้ายังไม่มี
if [ ! f "sync_log.txt" ]; then
  touch sync_log.txt
fi

if [ ! -f "config.yaml" ]; then
  cat > config.yaml <<EOF
database:
  host: 192.168.x.x
  port: 5432
  user: postgres
  password: yourpassword
api:
  key: your-api-key
  base_url: https://api.dedepos.com/
databases:
  - name: your-database-name
EOF
  echo "สร้าง config.yaml แล้ว กรุณาแก้ไขค่าให้ถูกต้องก่อนรันใหม่"
  exit 1
fi

# pull image ล่าสุดและรัน
docker compose pull
docker compose run --rm smlsynctodede
