# 🏥 Backend Appointment API - Go + Supabase

REST API สำหรับระบบจองนัดหมายตรวจสุขภาพ

## 🚀 Quick Start

### 1. ติดตั้ง Dependencies

```bash
go mod download
```

### 2. ตั้งค่า Environment Variables

สร้างไฟล์ `.env`:

```bash
cp .env.example .env
```

แก้ไขค่าใน `.env`:
- `SUPABASE_URL`: URL จาก Supabase project
- `SUPABASE_SERVICE_ROLE_KEY`: Service role key จาก Supabase
- `JWT_SECRET`: Secret key สำหรับ JWT (สุ่มเอง)

### 3. รันโปรเจค

```bash
go run main.go
```

Server จะรันที่ `http://localhost:8080`

## 📋 API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login ด้วยเบอร์โทร |
| POST | `/api/v1/auth/register` | สมัครสมาชิก |
| GET | `/api/v1/auth/me` | ดูข้อมูลตัวเอง (Auth) |
| PUT | `/api/v1/auth/me` | แก้ไขข้อมูลตัวเอง (Auth) |

### Bookings (Customer)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/bookings` | ดูการจองของตัวเอง |
| POST | `/api/v1/bookings` | สร้างการจอง |
| GET | `/api/v1/bookings/:id` | ดูรายละเอียดการจอง |
| PUT | `/api/v1/bookings/:id` | แก้ไขการจอง |
| DELETE | `/api/v1/bookings/:id` | ยกเลิกการจอง |

### Doctors & Schedules

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/doctors` | ดูรายชื่อแพทย์ |
| GET | `/api/v1/doctors/:id` | ดูข้อมูลแพทย์ |
| GET | `/api/v1/schedules` | ดูตารางเวลา |
| GET | `/api/v1/time-slots` | ดู time slots |
| GET | `/api/v1/time-slots/available` | ดู slots ว่าง |

### Nurse (Admin)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/nurse/bookings` | ดูการจองทั้งหมด |
| POST | `/api/v1/nurse/bookings` | สร้างการจองให้ลูกค้า |
| GET | `/api/v1/nurse/dashboard` | Dashboard |
| POST | `/api/v1/nurse/slots/block` | ตัด slot |

## 🔐 Authentication

ใช้ JWT Bearer Token:

```
Authorization: Bearer <token>
```

## 📦 Project Structure

```
backend-appointment/
├── config/          # Configuration & middleware
├── handlers/        # HTTP handlers
├── middleware/      # Auth middleware
├── models/          # Data models
├── routes/          # Route definitions
├── main.go          # Entry point
├── go.mod           # Dependencies
└── .env             # Environment variables
```

## 🛠️ Technologies

- **Go 1.21+**
- **Gin** - Web framework
- **Supabase** - Database & Auth
- **JWT** - Authentication
- **godotenv** - Environment variables

## 📝 License

MIT
