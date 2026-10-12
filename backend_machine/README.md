# Dashboard Produktivitas Mesin Sewing

Project Go sudah dipisah agar gampang dibaca dan deploy.

## Struktur

- `main.go` = entry point server
- `config/` = koneksi SQL Server dan environment variable
- `models/` = struct response API
- `repository/` = query SQL Server
- `handlers/` = endpoint API
- `utils/` = helper format, status, filter unicode
- `public/` = dashboard HTML

## Jalankan lokal

```powershell
cd C:\rexy\backend_machine
go mod tidy
go run .
```

Buka:

```text
http://localhost:8080
http://localhost:8080/process.html
http://localhost:8080/api/health
```

## API produktivitas rentang tanggal

`GET /api/productivity/range?start_date=2026-10-01&end_date=2026-10-10&shift=ALL`

- `start_date` dan `end_date` wajib menggunakan format `YYYY-MM-DD`.
- Kedua tanggal termasuk dalam hasil; semua hari kalender termasuk Minggu diambil.
- Maksimal 31 hari per request. Tanggal terbalik atau tidak valid menghasilkan HTTP 400.
- `shift` opsional, default `ALL`; perhitungan mengikuti API produktivitas harian.
- Respons: `{ "start_date": "2026-10-01", "end_date": "2026-10-10", "total_days": 10, "days": [...] }`.
- Setiap elemen `days` berisi `summary` dan `rows` seperti `/api/productivity`; tanggal terdapat pada `summary.date`.
- Data diambil berurutan dengan timeout total 5 menit. Kesalahan fatal menghasilkan HTTP 500 tanpa respons parsial. Seperti API harian, mesin yang gagal dihitung dapat dilewati.
- Frontend dashboard dan export Excel masih menggunakan API harian.

## Environment variable opsional

```powershell
$env:DB_SERVER="10.5.0.107"
$env:DB_PORT="1433"
$env:DB_USER="sa"
$env:DB_PASSWORD="Satu1234!"
# Database project ini dikunci ke sewingiot di config/db.go
# (abaikan User env DB_NAME=lectra). Override opsional: SEWINGIOT_DB_NAME
```
