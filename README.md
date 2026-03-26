# smlsynctodede

เครื่องมือ sync ข้อมูลจาก **PostgreSQL** ไปยัง **DedePOS API**
ข้อมูลที่ sync: ซัพพลายเออร์ (AP), ลูกค้า (AR), สินค้า/บาร์โค้ด (IC)

---

## Development

### รันโดยตรง

```bash
go run cmd/main.go
```

### Build Windows 64-bit

```bash
GOOS=windows GOARCH=amd64 go build -o build/smlsynctodede.exe cmd/main.go
```

### Build Linux 64-bit

```bash
GOOS=linux GOARCH=amd64 go build -o build/smlsynctodede cmd/main.go
```

---

## Config (config.yaml)

```yaml
database:
  host: 192.168.x.x # IP เครื่อง PostgreSQL
  port: 5432
  user: postgres
  password: xxxxx
api:
  key: xxxxx # API Key ของ DedePOS
  base_url: https://api.dedepos.com/
databases:
  - name: ชื่อ_database # ชื่อ database ที่ต้องการ sync
```

---

## Deploy ด้วย Docker (แนะนำสำหรับ Server)

### ภาพรวม

Image ถูก build และ push ขึ้น GitHub Container Registry อัตโนมัติทุกครั้งที่ push code ขึ้น `main`

```text
Push code → GitHub Actions → build image → ghcr.io/smlsoft/smlsynctodede:latest
```

ทีมที่ติดตั้งบน server ไม่ต้องมี source code ไม่ต้อง login แค่มี Docker ก็รันได้เลย

### โฟลเดอร์ที่ส่งให้ทีม

```text
docker/
├── docker-compose.yml ← config การรัน
├── config.yaml        ← แก้ให้ตรงกับ server ลูกค้า
└── sync_log.txt       ← ไฟล์เปล่าสำหรับเก็บ log
```

### ติดตั้งบน Server (ทีมที่รับไป)

#### ขั้นที่ 1 — copy โฟลเดอร์ docker/ ขึ้น server

```bash
ssh user@server "mkdir -p ~/data/smlsynctodede"
scp -r docker/ user@server:~/data/smlsynctodede
```

#### ขั้นที่ 2 — แก้ config ให้ตรงกับ server ลูกค้า

```bash
cd ~/data/smlsynctodede
nano config.yaml
```

#### ขั้นที่ 3 — รัน sync

```bash
docker compose run --rm smlsynctodede
```

Docker จะ pull image จาก `ghcr.io/smlsoft/smlsynctodede:latest` อัตโนมัติครับ

#### ขั้นที่ 4 — ดู log

```bash
cat sync_log.txt
```

### อัปเดตโปรแกรม

Push code ขึ้น GitHub ตามปกติ GitHub Actions จะ build image ใหม่ให้อัตโนมัติ

บน server รันเพื่อดึง image ใหม่:

```bash
docker compose pull
docker compose run --rm smlsynctodede
```

---

## หมายเหตุ

| ข้อ                            | รายละเอียด                             |
| ------------------------------ | -------------------------------------- |
| Server ต้องติดตั้ง Docker ก่อน | `docker` และ `docker compose`          |
| container ลบตัวเองหลังรัน      | เพราะใช้ `--rm`                        |
| log เก็บไว้ใน `sync_log.txt`   | ดูได้หลังรันเสร็จ                      |
| image                          | `ghcr.io/smlsoft/smlsynctodede:latest` |
