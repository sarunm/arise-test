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

**Unit Tests**

| Test | สิ่งที่ verify |
|---|---|
| `GET /customers/:id` — `TestGetByID_CacheHit` | คืนข้อมูลจาก cache โดยไม่แตะ DB |
| `GET /customers/:id` — `TestGetByID_CacheMiss` | ดึงจาก DB แล้ว set cache |
| `GET /customers/:id` — `TestGetByID_NotFound` | คืน error `ErrCustomerNotFound` |
| `POST /customers` — `TestCreateCustomer_Success` | สร้างลูกค้าสำเร็จ คืน id ที่ DB generate ให้ |
| `POST /customers` — `TestCreateCustomer_EmailAlreadyExists` | คืน `ErrEmailAlreadyExists` ไม่เรียก Create |
| `POST /customers` — `TestCreateCustomer_GetByEmailError` | DB error ระหว่างเช็ค email → คืน error ไม่เรียก Create |
| `POST /customers` — `TestCreateCustomer_RepoError` | DB error ระหว่าง insert → คืน error |
| `PATCH /customers/:id` — `TestUpdateCustomer_Success` | อัปเดตสำเร็จ invalidate cache |
| `PATCH /customers/:id` — `TestUpdateCustomer_PartialUpdate` | ส่งมาแค่ field เดียว field อื่นไม่เปลี่ยน |
| `PATCH /customers/:id` — `TestUpdateCustomer_NotFound` | คืน `ErrCustomerNotFound` ไม่ invalidate cache |
| `PATCH /customers/:id` — `TestUpdateCustomer_UpdateError` | DB error → คืน error ไม่ invalidate cache |

### Accounts

```
POST   /accounts                          เปิดบัญชีใหม่
GET    /accounts                          ดูบัญชีทั้งหมด
GET    /accounts/:id                      ดูข้อมูลบัญชี
GET    /accounts/customer/:customer_id    ดูบัญชีทั้งหมดของลูกค้า
```

```bash
# เปิดบัญชีใหม่
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"customer_id":1}'

# ดูบัญชีทั้งหมด
curl http://localhost:8080/api/v1/accounts

# ดูข้อมูลบัญชี
curl http://localhost:8080/api/v1/accounts/1

# ดูบัญชีทั้งหมดของลูกค้า
curl http://localhost:8080/api/v1/accounts/customer/1
```

**Unit Tests**

| Test | สิ่งที่ verify |
|---|---|
| `GET /accounts` — `TestGetAll_CacheHit` | คืน list จาก cache โดยไม่แตะ DB |
| `GET /accounts` — `TestGetAll_CacheMiss` | ดึงจาก DB แล้ว set cache |
| `GET /accounts` — `TestGetAll_RepoError` | DB error → คืน error |
| `GET /accounts/:id` — `TestGetByID_CacheHit` | คืนข้อมูลจาก cache โดยไม่แตะ DB |
| `GET /accounts/:id` — `TestGetByID_CacheMiss` | ดึงจาก DB แล้ว set cache |
| `GET /accounts/:id` — `TestGetByID_NotFound` | คืน error `ErrAccountNotFound` |
| `GET /accounts/customer/:id` — `TestGetByCustomerID_CacheHit` | คืน list จาก cache โดยไม่แตะ DB |
| `GET /accounts/customer/:id` — `TestGetByCustomerID_CacheMiss` | ดึงจาก DB แล้ว set cache |
| `GET /accounts/customer/:id` — `TestGetByCustomerID_RepoError` | DB error → คืน error |
| `POST /accounts` — `TestCreateAccount_Success` | สร้างบัญชีสำเร็จ status เป็น ACTIVE invalidate list cache + all cache |
| `POST /accounts` — `TestCreateAccount_RepoError` | DB error → คืน error ไม่ invalidate cache |

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

**Unit Tests**

| Test | สิ่งที่ verify |
|---|---|
| `POST /deposit` — `TestDeposit_Success` | ฝากเงินสำเร็จ invalidate tx cache + account cache |
| `POST /deposit` — `TestDeposit_RepoError` | คืน `ErrAccountNotActive` ไม่ invalidate cache |
| `POST /withdraw` — `TestWithdraw_Success` | ถอนเงินสำเร็จ invalidate tx cache + account cache |
| `POST /withdraw` — `TestWithdraw_InsufficientBalance` | คืน `ErrInsufficientBalance` ไม่ invalidate cache |
| `POST /transfer` — `TestTransfer_Success` | โอนเงินสำเร็จ invalidate cache ทั้ง 2 บัญชีพร้อมกัน |
| `POST /transfer` — `TestTransfer_SameAccount` | คืน `ErrSameAccount` ไม่แตะ DB เลย |
| `POST /transfer` — `TestTransfer_InsufficientBalance` | คืน `ErrInsufficientBalance` ไม่ invalidate cache |
| `GET /account/:id` — `TestGetByAccountID_CacheHit` | คืน list จาก cache โดยไม่แตะ DB |
| `GET /account/:id` — `TestGetByAccountID_CacheMiss` | ดึงจาก DB แล้ว set cache |
| `GET /account/:id` — `TestGetByAccountID_RepoError` | DB error → คืน error |

**Error responses**

| Status | เกิดจาก |
|---|---|
| `400` | Request body ไม่ถูกต้อง |
| `404` | ไม่พบลูกค้า / บัญชี |
| `409` | Email ซ้ำ |
| `422` | ยอดเงินไม่พอ, บัญชีไม่ active, โอนหาตัวเอง |
| `500` | Internal server error |
