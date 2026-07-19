-- Sesuaikan nama database sebelum menjalankan script ini.
-- USE NamaDatabaseAnda;
-- GO

IF COL_LENGTH('dbo.employee', 'id_telegram') IS NULL
BEGIN
    ALTER TABLE dbo.employee
    ADD id_telegram BIGINT NULL;
END
ELSE
BEGIN
    ALTER TABLE dbo.employee
    ALTER COLUMN id_telegram BIGINT NULL;
END;
GO

-- NIK harus unik agar registrasi tidak ambigu.
SELECT
    nik,
    COUNT(*) AS jumlah
FROM dbo.employee
WHERE nik IS NOT NULL
GROUP BY nik
HAVING COUNT(*) > 1;
GO

-- Satu Telegram ID hanya boleh terhubung ke satu karyawan.
IF NOT EXISTS (
    SELECT 1
    FROM sys.indexes
    WHERE name = 'UX_employee_id_telegram'
      AND object_id = OBJECT_ID('dbo.employee')
)
BEGIN
    CREATE UNIQUE INDEX UX_employee_id_telegram
        ON dbo.employee(id_telegram)
        WHERE id_telegram IS NOT NULL;
END;
GO

SELECT
    nik,
    name,
    branchdetail,
    id_telegram
FROM dbo.employee
ORDER BY nik;
GO
