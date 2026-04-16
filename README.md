# Description

REST API สำหรับระบบธนาคารเบื้องต้น รองรับการจัดการลูกค้า บัญชี และธุรกรรมทางการเงิน

## Tech Stack

- **Go 1.25** + Gin
- **PostgreSQL 17** — จัดเก็บข้อมูลหลัก
- **Redis 7** — Cache เพื่อลด DB load
- **GORM** — ORM
- **Docker Compose** — รัน services ทั้งหมดด้วยคำสั่งเดียว

## วิธีรัน

```bash
cp .env.example .env
docker compose up --build
```

App จะพร้อมที่ `http://localhost:8080`

### รัน Unit Test

```bash
go test ./... -cover
```

---

## API

Base path: `/api/v1`

### Customers

```
POST   /customers           สร้างลูกค้าใหม่
GET    /customers/:id       ดูข้อมูลลูกค้า
PATCH  /customers/:id       แก้ไขข้อมูลลูกค้า (ส่งมาแค่ field ที่อยากเปลี่ยน)
```

```bash
# สร้างลูกค้าใหม่
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"first_name":"สมชาย","last_name":"ใจดี","email":"somchai@example.com"}'

# ดูข้อมูลลูกค้า
curl http://localhost:8080/api/v1/customers/1

# แก้ไขข้อมูล (ส่งมาแค่ field ที่อยากเปลี่ยน)
curl -X PATCH http://localhost:8080/api/v1/customers/1 \
  -H "Content-Type: application/json" \
  -d '{"first_name":"สมหญิง"}'
```

### Accounts

```
POST   /accounts                          เปิดบัญชีใหม่
GET    /accounts/:id                      ดูข้อมูลบัญชี
GET    /accounts/customer/:customer_id    ดูบัญชีทั้งหมดของลูกค้า
```

```bash
# เปิดบัญชีใหม่
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"customer_id":1}'

# ดูข้อมูลบัญชี
curl http://localhost:8080/api/v1/accounts/1

# ดูบัญชีทั้งหมดของลูกค้า
curl http://localhost:8080/api/v1/accounts/customer/1
```

### Transactions

```
POST   /transactions/deposit              ฝากเงิน
POST   /transactions/withdraw             ถอนเงิน
POST   /transactions/transfer             โอนเงิน
GET    /transactions/account/:account_id  ดูประวัติธุรกรรม
```

> `amount` หน่วยเป็น **สตางค์** (100000 = 1,000.00 บาท) เพื่อหลีกเลี่ยงปัญหา floating-point

```bash
# ฝากเงิน 1,000 บาท
curl -X POST http://localhost:8080/api/v1/transactions/deposit \
  -H "Content-Type: application/json" \
  -d '{"account_id":1,"amount":100000}'

# ถอนเงิน 500 บาท
curl -X POST http://localhost:8080/api/v1/transactions/withdraw \
  -H "Content-Type: application/json" \
  -d '{"account_id":1,"amount":50000}'

# โอนเงิน 200 บาท จากบัญชี 1 ไปบัญชี 2
curl -X POST http://localhost:8080/api/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -d '{"from_account_id":1,"to_account_id":2,"amount":20000}'

# ดูประวัติธุรกรรมของบัญชี
curl http://localhost:8080/api/v1/transactions/account/1
```

**Error responses**

| Status | เกิดจาก |
|---|---|
| `400` | Request body ไม่ถูกต้อง |
| `404` | ไม่พบลูกค้า / บัญชี |
| `409` | Email ซ้ำ |
| `422` | ยอดเงินไม่พอ, บัญชีไม่ active, โอนหาตัวเอง |
| `500` | Internal server error |
