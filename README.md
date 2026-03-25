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
  host: 192.168.x.x     # IP เครื่อง PostgreSQL
  port: 5432
  user: postgres
  password: xxxxx
api:
  key: xxxxx             # API Key ของ DedePOS
  base_url: https://api.dedepos.com/
databases:
  - name: ชื่อ_database  # ชื่อ database ที่ต้องการ sync
```

---

## Deploy ด้วย Docker (แนะนำสำหรับ Server)

### ภาพรวม

```text
เครื่อง Dev                          Server ลูกค้า
─────────────────                    ─────────────────
docker build
    ↓
docker save → smlsynctodede.tar ───→ docker load
                                         ↓
                                     docker compose run
                                         ↓
                                     sync ข้อมูล → API
```

### โฟลเดอร์ที่ส่งให้ทีม

```text
docker/
├── smlsynctodede.tar  ← image โปรแกรมทั้งหมด
├── docker-compose.yml ← config การรัน
├── config.yaml        ← แก้ให้ตรงกับ server ลูกค้า
└── sync_log.txt       ← ไฟล์เปล่าสำหรับเก็บ log
```

### Build และ Export image (ทำบนเครื่อง Dev)

```bash
# 1. build image
docker build -f docker/Dockerfile -t smlsynctodede:latest .

# 2. export เป็นไฟล์ .tar
docker save smlsynctodede:latest -o docker/smlsynctodede.tar
```

### ติดตั้งบน Server (ทีมที่รับไป)

#### ขั้นที่ 1 — copy โฟลเดอร์ docker/ ขึ้น server

```bash
scp -r docker/ user@server:/opt/smlsynctodede
```

#### ขั้นที่ 2 — โหลด image เข้า Docker (ครั้งแรกครั้งเดียว)

```bash
cd /opt/smlsynctodede
docker load -i smlsynctodede.tar
```

#### ขั้นที่ 3 — แก้ config ให้ตรงกับ server ลูกค้า

```bash
nano config.yaml
```

#### ขั้นที่ 4 — รัน sync

```bash
docker compose run --rm smlsynctodede
```

#### ขั้นที่ 5 — ดู log

```bash
cat sync_log.txt
```

### อัปเดตโปรแกรม

```bash
# บนเครื่อง Dev: build และ export ใหม่
docker build -f docker/Dockerfile -t smlsynctodede:latest .
docker save smlsynctodede:latest -o docker/smlsynctodede.tar

# บน Server: โหลด image ใหม่
docker load -i smlsynctodede.tar
```

---

## หมายเหตุ

| ข้อ | รายละเอียด |
| --- | --- |
| Server ต้องติดตั้ง Docker ก่อน | `docker` และ `docker compose` |
| container ลบตัวเองหลังรัน | เพราะใช้ `--rm` |
| log เก็บไว้ใน `sync_log.txt` | ดูได้หลังรันเสร็จ |
| image size | ~8MB |
