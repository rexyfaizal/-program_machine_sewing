# Telegram Notification - Registrasi ID Karyawan

Program ini menyimpan Telegram User ID ke kolom `id_telegram` pada tabel SQL Server berdasarkan NIK karyawan.

## Struktur

```text
telegram_notif/
├── main.go
├── bot/
│   └── poller.go
├── config/
│   └── config.go
├── database/
│   └── sqlserver.go
├── handler/
│   └── telegram_handler.go
├── models/
│   ├── employee.go
│   └── telegram.go
├── repository/
│   └── employee_repository.go
├── service/
│   └── registration_service.go
├── telegram/
│   └── client.go
├── scripts/
│   └── setup_database.sql
├── build.ps1
├── run_bot.ps1.example
├── go.mod
└── .gitignore
```

## Tanggung jawab folder

- `models`: struktur data employee, hasil registrasi, dan payload Telegram.
- `repository`: seluruh query dan transaksi SQL Server.
- `service`: validasi NIK dan aturan bisnis registrasi.
- `handler`: menangani `/start`, `/cancel`, dan input NIK.
- `telegram`: client HTTP Telegram Bot API.
- `bot`: proses long polling `getUpdates`.
- `config`: membaca environment variable.
- `database`: membuka dan menguji koneksi SQL Server.
- `main.go`: menyambungkan seluruh komponen.

## Persiapan

1. Jalankan `scripts/setup_database.sql` melalui SQL Server Management Studio.
2. Salin `run_bot.ps1.example` menjadi `run_bot.ps1`.
3. Isi token Telegram dan konfigurasi SQL Server.

## Menjalankan untuk development

```powershell
go mod tidy
powershell -ExecutionPolicy Bypass -File .\run_bot.ps1
```

## Build untuk deployment Windows

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

Hasil build berada di folder `dist`.

Pada komputer server, cukup salin:

```text
telegram_notif.exe
run_bot.ps1
```

Kemudian jalankan:

```powershell
powershell -ExecutionPolicy Bypass -File .\run_bot.ps1
```

## Pengujian

Kirim pesan berikut ke bot:

```text
/start
```

Setelah bot meminta NIK, kirim contoh:

```text
100009
```

Bot juga mendukung:

```text
/start 100009
```

Periksa database:

```sql
SELECT nik, name, branchdetail, id_telegram
FROM dbo.employee
WHERE nik = '100009';
```

## Keamanan

- Jangan commit `run_bot.ps1` karena berisi token dan password.
- Satu Telegram ID tidak dapat digunakan untuk lebih dari satu NIK.
- NIK yang sudah terhubung ke akun lain tidak ditimpa otomatis.
- Query input menggunakan parameter SQL.
