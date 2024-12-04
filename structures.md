project-root/
│
├── cmd/
│   └── main.go                 # Entry point aplikasi
│
├── config/                     # Konfigurasi environment
│   ├── config.go               # Konfigurasi umum
│   └── env.go                  # Pembacaan environment variables
│
├── internal/                   # Kode spesifik aplikasi
│   ├── models/                 # Definisi struktur data
│   │   ├── user.go
│   │   └── ...
│   │
│   ├── repositories/           # Akses data dari database
│   │   ├── user_repository.go
│   │   └── ...
│   │
│   ├── services/               # Logika bisnis
│   │   ├── user_service.go
│   │   └── ...
│   │
│   └── controllers/            # Handler HTTP
│       ├── user_controller.go
│       └── ...
│
├── pkg/                        # Package yang bisa digunakan ulang
│   ├── database/               # Koneksi database
│   │   └── postgres.go
│   │
│   ├── storage/                # Klien penyimpanan (minio)
│   │   └── minio.go
│   │
│   └── utils/                  # Utilitas umum
│       ├── validator.go
│       └── ...
│
├── routes/                     # Definisi rute
│   └── router.go
│
├── migrations/                 # Migrasi database
│
├── docs/                       # Dokumentasi API
│
├── go.mod
└── go.sum