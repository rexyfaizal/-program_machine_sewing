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

## Chat berdasarkan tanggal dari database

`POST /api/ai/chat` membaca tanggal pada pertanyaan tentang mesin/produktivitas dan mengambil data SQL Server secara langsung menggunakan perhitungan API harian. Contoh body:

```json
{"question":"tanggal 9 Oktober 2026 apa mesin yang paling produktif", "date":"2026-10-12"}
```

Tanggal dalam pertanyaan mengalahkan tanggal dashboard. Format yang didukung: `9 Oktober 2026`, `2026-10-09`, `09/10/2026`, `09-10-2026`, `hari ini`, dan `kemarin` (WIB). Satu tanggal per pertanyaan; tanggal tidak valid, tidak lengkap, atau banyak tanggal menghasilkan HTTP 400.

Filter `location`, `status`, dan shift tertentu dari dashboard tetap diterapkan. Shift kosong/`CURRENT` menjadi `ALL` untuk pertanyaan bertanggal. Insight dashboard dan status mesin live tidak dipakai sebagai data historis. Data operator historis belum dimuat pada jalur ini. Tanpa tanggal dalam pertanyaan, chat tetap memakai konteks dashboard.

Ranking menyebutkan tanggal dan sumber data, default satu mesin (atau lima jika memakai kata `top`); gunakan `top 5` atau `5 mesin` untuk meminta beberapa mesin. Frasa `top mesin terbaik` dihitung sebagai ranking produktivitas oleh backend, tanpa meminta model menyusun angka. Pertanyaan lanjutan seperti `kalo hari Kamis tanggal 8 Oktober 2026?` meneruskan jenis ranking dan jumlah mesin dari pertanyaan user sebelumnya di `history`, lalu mengambil ulang data tanggal baru. Jika query database gagal, chat mengembalikan error tanpa memakai data dashboard sebagai pengganti. Jika tidak ada aktivitas, backend menyatakan data tidak tersedia. Perhitungan mesin yang gagal tetap dapat dilewati sebagaimana API harian.

## Environment variable opsional

Untuk Docker Compose, backend memakai Ollama di `http://10.5.0.101:11434` secara default. Override melalui environment shell atau `.env` di root project (sejajar `docker-compose.yml`):

```dotenv
AI_PROVIDER=ollama
OLLAMA_BASE_URL=http://10.5.0.101:11434
OLLAMA_MODEL=llama3.2
```

`.env` dalam `backend_machine/` tidak otomatis dimuat oleh Compose atau disalin ke image runtime. Setelah mengganti konfigurasi, jalankan `docker compose up -d --build backend` dari root project. Periksa alamat yang aktif dengan `docker compose exec backend printenv OLLAMA_BASE_URL`, lalu koneksi dari container dengan `docker compose exec backend wget -qO- http://10.5.0.101:11434/api/tags`.

```powershell
$env:DB_SERVER="10.5.0.107"
$env:DB_PORT="1433"
$env:DB_USER="sa"
$env:DB_PASSWORD="Satu1234!"
# Database project ini dikunci ke sewingiot di config/db.go
# (abaikan User env DB_NAME=lectra). Override opsional: SEWINGIOT_DB_NAME
```
