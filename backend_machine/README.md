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

## Environment variable opsional

```powershell
$env:DB_SERVER="10.5.0.106"
$env:DB_PORT="1433"
$env:DB_USER="sa"
$env:DB_PASSWORD="satu1"
# Database project ini dikunci ke sewingiot di config/db.go
# (abaikan User env DB_NAME=lectra). Override opsional: SEWINGIOT_DB_NAME
```
