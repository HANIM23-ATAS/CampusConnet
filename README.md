# 🎓 CampusConnect

> **Üniversite etkinliklerini birbirine bağlayan modern mikroservis platformu.**

CampusConnect, öğrencilerin kampüs etkinliklerini keşfedip katılabileceği, gerçek zamanlı analizlerin Go servisi aracılığıyla takip edildiği tam yığın bir mikroservis platformudur. NestJS tabanlı ana API ve Go tabanlı analitik servisi Docker ile birlikte çalışır.

---

## 📐 Mimari Genel Bakış

```
┌─────────────────────────────────────────────────────────┐
│                      Docker Compose                     │
│                                                         │
│   ┌──────────────┐    Webhook    ┌──────────────────┐   │
│   │  NestJS API  │──────────────▶│   Go Analytics   │   │
│   │  :3000       │   HMAC-SHA256 │   Service :8080  │   │
│   └──────┬───────┘               └──────────────────┘   │
│          │ Prisma ORM                                    │
│   ┌──────▼───────┐                                      │
│   │  PostgreSQL  │                                      │
│   │  :5432       │                                      │
│   └──────────────┘                                      │
└─────────────────────────────────────────────────────────┘
```

---

## 🛠️ Teknoloji Yığını

| Servis | Teknoloji |
|--------|-----------|
| **API Servisi** | NestJS 10, TypeScript 5 |
| **Veritabanı ORM** | Prisma 5 + PostgreSQL 15 |
| **API Katmanı** | REST + GraphQL (Apollo) |
| **Kimlik Doğrulama** | JWT + bcrypt + Passport |
| **Dokümantasyon** | Swagger / OpenAPI |
| **Analitik Servisi** | Go (net/http) |
| **Güvenlik** | HMAC-SHA256 Webhook + API Key |
| **Rate Limiting** | Go `golang.org/x/time/rate` (60 req/min) |
| **Konteynerizasyon** | Docker + Docker Compose |

---

## 📁 Proje Yapısı

```
CampusConnect/
├── docker-compose.yml          # Tüm servislerin orkestrasyonu
├── requests.http               # API test dosyası (VS Code REST Client)
│
├── nestjs-service/             # Ana API Servisi
│   ├── Dockerfile
│   ├── start.sh                # Migrasyon + başlatma betiği
│   ├── package.json
│   ├── tsconfig.json
│   ├── prisma/
│   │   └── schema.prisma       # Veritabanı şeması
│   └── src/
│       ├── main.ts
│       ├── app.module.ts
│       ├── auth/               # JWT kimlik doğrulama
│       ├── users/              # Kullanıcı CRUD
│       ├── events/             # Etkinlik CRUD + filtreleme
│       ├── graphql/            # GraphQL resolver'ları
│       ├── webhooks/           # Go servisine webhook tetikleyici
│       └── prisma/             # Prisma servisi
│
└── go-service/                 # Analitik Servisi
    ├── Dockerfile
    ├── go.mod
    └── main.go                 # HTTP server, webhook + analitik API
```

---

## 🗄️ Veritabanı Şeması

```prisma
model User {
  id        Int                @id @default(autoincrement())
  email     String             @unique
  password  String             // bcrypt hash
  name      String
  events    EventParticipant[]
  createdAt DateTime           @default(now())
  updatedAt DateTime           @updatedAt
}

model Event {
  id           Int                @id @default(autoincrement())
  title        String
  description  String
  date         DateTime
  location     String
  category     String             @default("General")
  participants EventParticipant[]
  createdAt    DateTime           @default(now())
  updatedAt    DateTime           @updatedAt
}

model EventParticipant {
  id        Int      @id @default(autoincrement())
  eventId   Int
  userId    Int
  event     Event    @relation(...)
  user      User     @relation(...)
  createdAt DateTime @default(now())

  @@unique([eventId, userId])   // Bir kullanıcı etkinliğe bir kez katılabilir
}
```

---

## 🚀 Kurulum ve Çalıştırma

### Ön Gereksinimler

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) yüklü ve çalışıyor olmalı
- Git

### 1. Repoyu Klonla

```bash
git clone https://github.com/<your-username>/CampusConnect.git
cd CampusConnect
```

### 2. Servisleri Başlat

```bash
docker compose up --build
```

Bu komut şunları yapar:
- PostgreSQL veritabanını başlatır ve sağlık kontrolünü bekler
- Prisma migrasyonlarını otomatik olarak çalıştırır
- NestJS API'yi `http://localhost:3000` üzerinde başlatır
- Go Analitik Servisini `http://localhost:8080` üzerinde başlatır

### 3. Servisleri Durdur

```bash
docker compose down          # Servisleri durdur
docker compose down -v       # Servisleri durdur ve veriyi sil
```

---

## 📡 API Referansı

### 🔐 Kimlik Doğrulama (REST)

| Metot | Endpoint | Açıklama | Auth |
|-------|----------|----------|------|
| `POST` | `/users` | Kullanıcı kaydı | Hayır |
| `POST` | `/auth/login` | Giriş → JWT token | Hayır |

**Kayıt Örneği:**
```http
POST http://localhost:3000/users
Content-Type: application/json

{
  "email": "ogrenci@kampus.edu.tr",
  "password": "Sifre123!",
  "name": "Ali Yılmaz"
}
```

**Giriş Örneği:**
```http
POST http://localhost:3000/auth/login
Content-Type: application/json

{
  "email": "ogrenci@kampus.edu.tr",
  "password": "Sifre123!"
}
```

**Yanıt:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

### 📅 Etkinlikler (REST)

| Metot | Endpoint | Açıklama | Auth |
|-------|----------|----------|------|
| `POST` | `/events` | Etkinlik oluştur | ✅ JWT |
| `GET` | `/events` | Etkinlik listesi (sayfalama + filtre) | Hayır |

**Etkinlik Oluşturma:**
```http
POST http://localhost:3000/events
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Bahar Hackathon 2024",
  "description": "Teknoloji tutkunları için büyük etkinlik.",
  "date": "2024-05-20T10:00:00Z",
  "location": "İstanbul",
  "category": "Teknoloji"
}
```

**Filtreleyerek Listeleme:**
```http
GET http://localhost:3000/events?page=1&limit=10&location=İstanbul&category=Teknoloji
```

---

### 🔗 Etkinlik Katılımı (GraphQL)

GraphQL Playground: **`http://localhost:3000/graphql`**

**Etkinliğe Katıl:**
```graphql
mutation {
  joinEvent(eventId: 1) {
    id
    eventId
    userId
  }
}
```

**Etkinlikten Ayrıl:**
```graphql
mutation {
  leaveEvent(eventId: 1)
}
```

> ⚠️ GraphQL isteklerinde `Authorization: Bearer <token>` header'ı gereklidir.

---

### 📊 Analitik Servisi (Go — Port 8080)

Tüm analitik endpoint'leri `X-API-Key` header'ı gerektirir.

| Metot | Endpoint | Açıklama |
|-------|----------|----------|
| `POST` | `/webhook` | NestJS'den etkinlik verisi alır (HMAC doğrulamalı) |
| `GET` | `/analytics/top-events` | En popüler 5 etkinlik |
| `GET` | `/analytics/stats` | Son 7 günün katılım istatistikleri |

**Top Etkinlikler:**
```http
GET http://localhost:8080/analytics/top-events
X-API-Key: secure_api_key_123
```

**Yanıt:**
```json
{
  "topEvents": [
    { "eventId": 1, "count": 42 },
    { "eventId": 3, "count": 28 }
  ]
}
```

**Haftalık İstatistikler:**
```http
GET http://localhost:8080/analytics/stats
X-API-Key: secure_api_key_123
```

> 🛡️ Rate limit aşıldığında RFC 7807 Problem Details formatında yanıt döner (60 istek/dakika).

---

## 🔒 Güvenlik

### JWT Kimlik Doğrulama
- Şifreler `bcrypt` ile hashlenir
- Token korumalı endpoint'lere erişim için `Authorization: Bearer <token>` header'ı kullanılır

### HMAC-SHA256 Webhook İmzalama
NestJS, Go servisine webhook gönderirken isteği HMAC-SHA256 ile imzalar. Go servisi bu imzayı doğrular — paylaşılan sırrı bilmeyen istemciler webhook verisini enjekte edemez.

### API Key Koruması
Go Analitik Servisi'ndeki tüm analitik endpoint'leri `X-API-Key` header'ı ile korunur.

### Rate Limiting
Go servisi, dakikada 60 istek sınırı uygular. Limit aşıldığında standart RFC 7807 hata yanıtı döner:
```json
{
  "type": "about:blank",
  "title": "Too Many Requests",
  "status": 429,
  "detail": "Rate limit exceeded (60 req/min)."
}
```

---

## 🌐 Ortam Değişkenleri

### NestJS Servisi

| Değişken | Varsayılan | Açıklama |
|----------|-----------|----------|
| `DATABASE_URL` | `postgresql://campus_user:campus_password@postgres:5432/campusconnect` | PostgreSQL bağlantı URL'i |
| `JWT_SECRET` | `super_secret_jwt_key` | JWT imzalama anahtarı |
| `WEBHOOK_URL` | `http://go-service:8080/webhook` | Go servis webhook URL'i |
| `WEBHOOK_SECRET` | `my_webhook_secret` | HMAC imzalama sırrı |

### Go Servisi

| Değişken | Varsayılan | Açıklama |
|----------|-----------|----------|
| `WEBHOOK_SECRET` | `my_webhook_secret` | HMAC doğrulama sırrı |
| `API_KEY` | `secure_api_key_123` | Analitik endpoint API anahtarı |

> ⚠️ **Üretim ortamında** tüm değerleri güçlü, rastgele değerlerle değiştirin.

---

## 🧪 API Testleri

Projeyle birlikte gelen `requests.http` dosyasını kullanabilirsiniz.

**VS Code ile:**
1. [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) eklentisini yükleyin
2. `requests.http` dosyasını açın
3. Her isteğin üzerindeki `Send Request` bağlantısına tıklayın

**Önerilen Test Sırası:**
1. `POST /users` → Kullanıcı oluştur
2. `POST /auth/login` → Token al (otomatik kaydedilir)
3. `POST /events` → Etkinlik oluştur
4. `GET /events` → Filtreleyerek listele
5. GraphQL ile `joinEvent` → Etkinliğe katıl
6. `GET /analytics/top-events` → Analitiği görüntüle

---

## 📜 Lisans

Bu proje MIT lisansı altında dağıtılmaktadır. Daha fazla bilgi için [LICENSE](LICENSE) dosyasına bakın.

---

<div align="center">
  <strong>CampusConnect</strong> — Kampüste her etkinlik bir bağlantıdır. 🎓
</div>
