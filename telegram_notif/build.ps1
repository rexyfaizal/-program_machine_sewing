$ErrorActionPreference = 'Stop'

Write-Host 'Merapikan dan memeriksa project...'
go fmt ./...
go mod tidy
go test ./...

New-Item -ItemType Directory -Force -Path '.\dist' | Out-Null

go build `
    -trimpath `
    -ldflags='-s -w' `
    -o '.\dist\telegram_notif.exe' `
    .

Copy-Item '.\run_bot.ps1.example' '.\dist\run_bot.ps1.example' -Force
Copy-Item '.\scripts\setup_database.sql' '.\dist\setup_database.sql' -Force

Write-Host ''
Write-Host 'Build selesai:'
Write-Host '  dist\telegram_notif.exe'
Write-Host '  dist\run_bot.ps1.example'
Write-Host '  dist\setup_database.sql'
