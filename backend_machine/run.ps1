# Jalankan backend project machine sewing.
# Database dikunci ke sewingiot di config/db.go
# (tidak terpengaruh User env DB_NAME=lectra).

Write-Host "Starting backend_machine (database=sewingiot)..."
Set-Location $PSScriptRoot
go run .
