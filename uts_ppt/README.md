# Legal Consultation API

REST API backend untuk platform konsultasi hukum berbasis web, dibangun dengan **Go + Gin** menggunakan **Clean Architecture**.

## 🏗️ Struktur Proyek

```
legal-consultation-api/
├── cmd/api/main.go                  # Entrypoint
├── internal/
│   ├── config/config.go             # Konfigurasi aplikasi
│   ├── database/database.go         # Koneksi PostgreSQL
│   ├── models/models.go             # Domain models
│   ├── repository/                  # Layer data access
│   │   ├── user_repository.go
│   │   ├── lawyer_repository.go
│   │   ├── consultation_repository.go
│   │   ├── payment_repository.go
│   │   ├── chat_repository.go
│   │   └── review_schedule_repository.go
│   ├── service/                     # Business logic layer
│   │   ├── auth_service.go
│   │   ├── lawyer_service.go
│   │   ├── consultation_service.go
│   │   ├── payment_service.go
│   │   └── chat_review_schedule_service.go
│   ├── handler/                     # HTTP handlers
│   │   ├── auth_handler.go
│   │   ├── lawyer_handler.go
│   │   ├── consultation_handler.go
│   │   ├── payment_handler.go
│   │   └── chat_handler.go
│   ├── middleware/middleware.go      # Auth, CORS, Logging
│   └── router/router.go             # Route definitions
├── pkg/
│   ├── jwt/jwt.go                   # JWT utilities
│   └── response/response.go        # Standard response helpers
├── migrations/001_init_schema.sql   # Database migration
├── .env.example
└── go.mod
```

## 🚀 Cara Menjalankan

### 1. Prerequisites
- Go 1.21+
- PostgreSQL 14+

### 2. Setup Database
```bash
psql -U postgres -c "CREATE DATABASE legal_consultation_db;"
psql -U postgres -d legal_consultation_db -f migrations/001_init_schema.sql
```

### 3. Konfigurasi Environment
```bash
copy .env.example .env
# Edit .env sesuai konfigurasi lokal
```

### 4. Install Dependencies & Run
```bash
go mod tidy
go run ./cmd/api/main.go
```

Server berjalan di: `http://localhost:8080`

---

## 📋 API Endpoints

### Auth
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| POST | `/api/auth/register` | Daftar akun baru | ❌ |
| POST | `/api/auth/login` | Login | ❌ |
| GET | `/api/profile` | Lihat profil | ✅ |
| PUT | `/api/profile` | Update profil | ✅ |

### Lawyers
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/lawyers` | Cari lawyer (filter) | ❌ |
| GET | `/api/lawyers/:id` | Detail lawyer | ❌ |
| POST | `/api/lawyers/profile` | Buat profil lawyer | ✅ Lawyer |
| PUT | `/api/lawyers/profile` | Update profil lawyer | ✅ Lawyer |
| PATCH | `/api/lawyers/availability` | Set ketersediaan | ✅ Lawyer |
| GET | `/api/lawyers/:id/schedules` | Jadwal lawyer | ❌ |
| POST | `/api/schedules` | Tambah jadwal | ✅ Lawyer |
| DELETE | `/api/schedules/:id` | Hapus jadwal | ✅ Lawyer |
| GET | `/api/lawyers/:id/reviews` | Ulasan lawyer | ❌ |

### Consultations
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| POST | `/api/consultations` | Booking konsultasi | ✅ Client |
| GET | `/api/consultations` | Riwayat konsultasi | ✅ |
| GET | `/api/consultations/:id` | Detail konsultasi | ✅ |
| GET | `/api/consultations/:id/status` | Tracking status | ✅ |
| PATCH | `/api/consultations/:id/confirm` | Konfirmasi | ✅ Lawyer |
| PATCH | `/api/consultations/:id/cancel` | Batalkan | ✅ |
| PATCH | `/api/consultations/:id/complete` | Selesaikan | ✅ Lawyer |
| POST | `/api/consultations/:id/reviews` | Beri ulasan | ✅ Client |

### Payments
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/consultations/:id/payment` | Info pembayaran | ✅ |
| POST | `/api/payments/:id/upload` | Upload bukti bayar | ✅ |
| PATCH | `/api/admin/payments/:id/verify` | Verifikasi | ✅ Admin |
| PATCH | `/api/admin/payments/:id/reject` | Tolak | ✅ Admin |

### Chat
| Method | Endpoint | Deskripsi | Auth |
|--------|----------|-----------|------|
| GET | `/api/consultations/:id/messages` | Ambil pesan | ✅ |
| POST | `/api/consultations/:id/messages` | Kirim pesan (REST) | ✅ |
| GET | `/api/consultations/:id/ws` | WebSocket real-time | ✅ |

---

## 📦 Contoh Request/Response

### Register
```json
POST /api/auth/register
{
  "email": "client@example.com",
  "password": "Password123!",
  "full_name": "Budi Santoso",
  "phone": "081234567890",
  "role": "client"
}

Response 201:
{
  "status": "success",
  "message": "Registration successful",
  "data": {
    "user": { "id": "uuid", "email": "...", "role": "client" },
    "tokens": {
      "access_token": "eyJ...",
      "refresh_token": "eyJ...",
      "expires_at": 1713700000
    }
  }
}
```

### Search Lawyers
```
GET /api/lawyers?specialization=Pidana&city=Jakarta&min_rating=4&max_fee=500000&page=1&limit=10
```

### Book Consultation
```json
POST /api/consultations
Authorization: Bearer <token>
{
  "lawyer_id": "uuid-lawyer",
  "schedule_date": "2024-05-10",
  "start_time": "09:00",
  "end_time": "10:00",
  "duration_hours": 1,
  "case_description": "Saya membutuhkan konsultasi terkait sengketa tanah...",
  "case_type": "Perdata",
  "platform": "chat"
}
```

### Upload Payment Proof
```
POST /api/payments/:id/upload
Content-Type: multipart/form-data
Fields: proof (file), payment_method, bank_name, transfer_reference
```

### WebSocket Chat
```
ws://localhost:8080/api/consultations/:id/ws
Authorization: Bearer <token>

Send:  { "message_type": "text", "content": "Halo pak lawyer" }
Recv:  { "id": "uuid", "sender_id": "uuid", "content": "...", "created_at": "..." }
```
