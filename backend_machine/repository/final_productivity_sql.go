package repository

// finalProductivitySQL adalah rumus FINAL NORMAL + SHIFT (1:1 dengan query SSMS).
// Parameter:
//
//	@p_work_date, @p_uuid, @p_area, @p_line, @p_mode, @p_selected_shift, @p_process_table
const finalProductivitySQL = `
SET NOCOUNT ON;

DECLARE @work_date DATE = TRY_CONVERT(date, @p_work_date);
DECLARE @uuid NVARCHAR(100) = @p_uuid;
DECLARE @area NVARCHAR(50) = @p_area;
DECLARE @line NVARCHAR(255) = UPPER(LTRIM(RTRIM(@p_line)));
DECLARE @mode NVARCHAR(20) = @p_mode;
DECLARE @selected_shift NVARCHAR(50) = NULLIF(LTRIM(RTRIM(@p_selected_shift)), N'');
DECLARE @process_table SYSNAME = @p_process_table;

DECLARE @base DATETIME2(3) = CAST(@work_date AS DATETIME2(3));
DECLARE @server_now DATETIME2(3) = SYSDATETIME();

DROP TABLE IF EXISTS #groups;
CREATE TABLE #groups
(
    group_no INT NOT NULL,
    group_name NVARCHAR(50) NOT NULL,
    period_start DATETIME2(3) NOT NULL,
    period_end DATETIME2(3) NOT NULL,
    break_start DATETIME2(3) NULL,
    break_end DATETIME2(3) NULL
);

IF UPPER(@mode) = N'NORMAL'
BEGIN
    INSERT INTO #groups (group_no, group_name, period_start, period_end, break_start, break_end)
    VALUES (0, N'NORMAL', @base, DATEADD(DAY, 1, @base), NULL, NULL);
END;

IF UPPER(@mode) = N'SHIFT'
BEGIN
    -- Ada override khusus line? kalau ada pakai line, kalau tidak pakai default area (line_name = '').
    DECLARE @has_line_cfg BIT = 0;
    IF @line <> N'' AND EXISTS (
        SELECT 1 FROM dbo.shift_setting s
        WHERE UPPER(LTRIM(RTRIM(s.area))) = UPPER(LTRIM(RTRIM(@area)))
          AND UPPER(LTRIM(RTRIM(ISNULL(s.line_name, N'')))) = @line
          AND s.is_active = 1
          AND s.effective_from <= @work_date
          AND (s.effective_to IS NULL OR s.effective_to >= @work_date)
    )
        SET @has_line_cfg = 1;

    ;WITH cfg AS
    (
        SELECT
            s.shift_no,
            s.shift_name,
            s.start_time,
            s.end_time,
            s.break_start,
            s.break_end,
            DATEADD(SECOND, DATEDIFF(SECOND, CAST('00:00:00' AS TIME), s.start_time), @base) AS shift_start,
            DATEADD(
                DAY,
                CASE WHEN s.end_time <= s.start_time THEN 1 ELSE 0 END,
                DATEADD(SECOND, DATEDIFF(SECOND, CAST('00:00:00' AS TIME), s.end_time), @base)
            ) AS shift_end
        FROM dbo.shift_setting s
        WHERE UPPER(LTRIM(RTRIM(s.area))) = UPPER(LTRIM(RTRIM(@area)))
          AND s.is_active = 1
          AND s.effective_from <= @work_date
          AND (s.effective_to IS NULL OR s.effective_to >= @work_date)
          AND (
                (@has_line_cfg = 1 AND UPPER(LTRIM(RTRIM(ISNULL(s.line_name, N'')))) = @line)
             OR (@has_line_cfg = 0 AND ISNULL(LTRIM(RTRIM(s.line_name)), N'') = N'')
          )
          AND (@selected_shift IS NULL OR UPPER(LTRIM(RTRIM(s.shift_name))) = UPPER(LTRIM(RTRIM(@selected_shift))))
    ),
    cfg_break AS
    (
        SELECT
            *,
            CASE
                WHEN break_start IS NULL THEN NULL
                ELSE DATEADD(
                    DAY,
                    CASE WHEN end_time <= start_time AND break_start < start_time THEN 1 ELSE 0 END,
                    DATEADD(SECOND, DATEDIFF(SECOND, CAST('00:00:00' AS TIME), break_start), @base)
                )
            END AS break_start_dt,
            CASE
                WHEN break_end IS NULL THEN NULL
                ELSE DATEADD(
                    DAY,
                    CASE WHEN end_time <= start_time AND break_end < start_time THEN 1 ELSE 0 END,
                    DATEADD(SECOND, DATEDIFF(SECOND, CAST('00:00:00' AS TIME), break_end), @base)
                )
            END AS break_end_dt
        FROM cfg
    )
    INSERT INTO #groups (group_no, group_name, period_start, period_end, break_start, break_end)
    SELECT shift_no, shift_name, shift_start, shift_end, break_start_dt, break_end_dt
    FROM cfg_break;
END;

IF NOT EXISTS (SELECT 1 FROM #groups)
BEGIN
    RAISERROR(N'Setting shift tidak ditemukan.', 16, 1);
    RETURN;
END;

DROP TABLE IF EXISTS #segments;
CREATE TABLE #segments
(
    group_no INT NOT NULL,
    group_name NVARCHAR(50) NOT NULL,
    segment_no INT NOT NULL,
    segment_start DATETIME2(3) NOT NULL,
    segment_end DATETIME2(3) NOT NULL
);

INSERT INTO #segments (group_no, group_name, segment_no, segment_start, segment_end)
SELECT group_no, group_name, 1, period_start, period_end
FROM #groups
WHERE break_start IS NULL OR break_end IS NULL;

INSERT INTO #segments (group_no, group_name, segment_no, segment_start, segment_end)
SELECT group_no, group_name, 1, period_start, break_start
FROM #groups
WHERE break_start IS NOT NULL AND break_end IS NOT NULL AND break_start > period_start;

INSERT INTO #segments (group_no, group_name, segment_no, segment_start, segment_end)
SELECT group_no, group_name, 2, break_end, period_end
FROM #groups
WHERE break_start IS NOT NULL AND break_end IS NOT NULL AND break_end < period_end;

DECLARE @global_start DATETIME2(3);
DECLARE @global_end DATETIME2(3);
SELECT @global_start = MIN(period_start), @global_end = MAX(period_end) FROM #groups;

DECLARE @cutoff_time DATETIME2(3) =
    CASE
        WHEN @server_now <= @global_start THEN @global_start
        WHEN @server_now >= @global_end THEN @global_end
        ELSE @server_now
    END;

DECLARE @mac_state INT =
(
    SELECT TOP (1) TRY_CONVERT(INT, MacState)
    FROM dbo.MachineInfo
    WHERE LOWER(LTRIM(RTRIM(UUID))) = LOWER(LTRIM(RTRIM(@uuid)))
);

DROP TABLE IF EXISTS #process_raw;
CREATE TABLE #process_raw
(
    process_start DATETIME2(3) NOT NULL,
    process_seconds INT NOT NULL,
    file_name NVARCHAR(255) NULL
);

IF OBJECT_ID(N'dbo.' + @process_table, N'U') IS NOT NULL
BEGIN
    DECLARE @sql NVARCHAR(MAX) = N'
        INSERT INTO #process_raw (process_start, process_seconds, file_name)
        SELECT
            TRY_CONVERT(DATETIME2(3), StartTime),
            TRY_CONVERT(INT, ProcTime),
            TRY_CONVERT(NVARCHAR(255), FileName)
        FROM dbo.' + QUOTENAME(@process_table) + N'
        WHERE TRY_CONVERT(DATETIME2(3), StartTime) >= DATEADD(DAY, -1, @p_global_start)
          AND TRY_CONVERT(DATETIME2(3), StartTime) < @p_cutoff
          AND TRY_CONVERT(INT, ProcTime) > 0;';

    EXEC sp_executesql
        @sql,
        N'@p_global_start DATETIME2(3), @p_cutoff DATETIME2(3)',
        @p_global_start = @global_start,
        @p_cutoff = @cutoff_time;
END;

DECLARE @last_shut DATETIME2(3) =
(
    SELECT MAX(TRY_CONVERT(DATETIME2(3), ShutTime))
    FROM dbo.Record_RunTime
    WHERE LOWER(LTRIM(RTRIM(UUID))) = LOWER(LTRIM(RTRIM(@uuid)))
      AND TRY_CONVERT(DATETIME2(3), ShutTime) IS NOT NULL
      AND TRY_CONVERT(DATETIME2(3), ShutTime) > @global_start
      AND TRY_CONVERT(DATETIME2(3), ShutTime) <= @cutoff_time
);

DECLARE @first_active_process DATETIME2(3) =
(
    SELECT MIN(process_start)
    FROM #process_raw
    WHERE process_start > COALESCE(@last_shut, @global_start)
      AND process_start < @cutoff_time
);

DROP TABLE IF EXISTS #runtime_intervals;
CREATE TABLE #runtime_intervals
(
    start_time DATETIME2(3) NOT NULL,
    end_time DATETIME2(3) NOT NULL,
    source NVARCHAR(20) NOT NULL
);

INSERT INTO #runtime_intervals (start_time, end_time, source)
SELECT
    CASE
        WHEN TRY_CONVERT(DATETIME2(3), StartTime) < @global_start THEN @global_start
        ELSE TRY_CONVERT(DATETIME2(3), StartTime)
    END,
    CASE
        WHEN TRY_CONVERT(DATETIME2(3), ShutTime) > @cutoff_time THEN @cutoff_time
        ELSE TRY_CONVERT(DATETIME2(3), ShutTime)
    END,
    N'CLOSED'
FROM dbo.Record_RunTime
WHERE LOWER(LTRIM(RTRIM(UUID))) = LOWER(LTRIM(RTRIM(@uuid)))
  AND TRY_CONVERT(DATETIME2(3), StartTime) < @cutoff_time
  AND TRY_CONVERT(DATETIME2(3), ShutTime) IS NOT NULL
  AND TRY_CONVERT(DATETIME2(3), ShutTime) > @global_start;

IF (@server_now >= @global_start AND @server_now < @global_end AND @mac_state IN (1, 2))
BEGIN
    INSERT INTO #runtime_intervals (start_time, end_time, source)
    SELECT
        CASE
            WHEN TRY_CONVERT(DATETIME2(3), StartTime) < @global_start THEN @global_start
            ELSE TRY_CONVERT(DATETIME2(3), StartTime)
        END,
        @cutoff_time,
        N'OPEN'
    FROM dbo.Record_RunTime
    WHERE LOWER(LTRIM(RTRIM(UUID))) = LOWER(LTRIM(RTRIM(@uuid)))
      AND ShutTime IS NULL
      AND TRY_CONVERT(DATETIME2(3), StartTime) < @cutoff_time;
END;

IF (
    @server_now >= @global_start
    AND @server_now < @global_end
    AND @mac_state IN (1, 2)
    AND @first_active_process IS NOT NULL
)
BEGIN
    INSERT INTO #runtime_intervals (start_time, end_time, source)
    VALUES (@first_active_process, @cutoff_time, N'FALLBACK');
END;

DELETE FROM #runtime_intervals WHERE end_time <= start_time;

DROP TABLE IF EXISTS #runtime_merged;
CREATE TABLE #runtime_merged
(
    start_time DATETIME2(3) NOT NULL,
    end_time DATETIME2(3) NOT NULL
);

;WITH ordered AS
(
    SELECT
        start_time,
        end_time,
        MAX(end_time) OVER (
            ORDER BY start_time, end_time
            ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING
        ) AS previous_max_end
    FROM #runtime_intervals
),
marked AS
(
    SELECT
        start_time,
        end_time,
        CASE
            WHEN previous_max_end IS NULL OR start_time > previous_max_end THEN 1
            ELSE 0
        END AS new_group
    FROM ordered
),
grouped AS
(
    SELECT
        start_time,
        end_time,
        SUM(new_group) OVER (
            ORDER BY start_time, end_time
            ROWS UNBOUNDED PRECEDING
        ) AS group_id
    FROM marked
)
INSERT INTO #runtime_merged (start_time, end_time)
SELECT MIN(start_time), MAX(end_time)
FROM grouped
GROUP BY group_id;

DROP TABLE IF EXISTS #process_intervals;
CREATE TABLE #process_intervals
(
    process_start DATETIME2(3) NOT NULL,
    process_end DATETIME2(3) NOT NULL
);

;WITH ordered_process AS
(
    SELECT
        process_start,
        process_seconds,
        LEAD(process_start) OVER (ORDER BY process_start, file_name) AS next_start
    FROM #process_raw
),
candidate AS
(
    SELECT
        process_start,
        next_start,
        CASE
            WHEN DATEADD(SECOND, process_seconds, process_start) > @cutoff_time THEN @cutoff_time
            ELSE DATEADD(SECOND, process_seconds, process_start)
        END AS candidate_end
    FROM ordered_process
),
final_process AS
(
    SELECT
        process_start,
        CASE
            WHEN next_start IS NOT NULL AND next_start < candidate_end THEN next_start
            ELSE candidate_end
        END AS process_end
    FROM candidate
)
INSERT INTO #process_intervals (process_start, process_end)
SELECT process_start, process_end
FROM final_process
WHERE process_end > process_start
  AND process_start < @cutoff_time
  AND process_end > @global_start;

DROP TABLE IF EXISTS #power_totals;
CREATE TABLE #power_totals
(
    group_no INT PRIMARY KEY,
    power_seconds BIGINT NOT NULL
);

INSERT INTO #power_totals (group_no, power_seconds)
SELECT
    s.group_no,
    SUM(DATEDIFF(SECOND, x.overlap_start, x.overlap_end))
FROM #segments s
INNER JOIN #runtime_merged r
    ON r.start_time < s.segment_end
   AND r.end_time > s.segment_start
CROSS APPLY
(
    SELECT
        CASE WHEN r.start_time > s.segment_start THEN r.start_time ELSE s.segment_start END AS overlap_start,
        CASE WHEN r.end_time < s.segment_end THEN r.end_time ELSE s.segment_end END AS overlap_end
) x
WHERE x.overlap_end > x.overlap_start
GROUP BY s.group_no;

DROP TABLE IF EXISTS #process_totals;
CREATE TABLE #process_totals
(
    group_no INT PRIMARY KEY,
    process_seconds BIGINT NOT NULL
);

INSERT INTO #process_totals (group_no, process_seconds)
SELECT
    s.group_no,
    SUM(DATEDIFF(SECOND, a.overlap_start, b.overlap_end))
FROM #segments s
INNER JOIN #runtime_merged r
    ON r.start_time < s.segment_end
   AND r.end_time > s.segment_start
INNER JOIN #process_intervals p
    ON p.process_start < s.segment_end
   AND p.process_end > s.segment_start
   AND p.process_start < r.end_time
   AND p.process_end > r.start_time
CROSS APPLY
(
    SELECT MAX(v) AS overlap_start
    FROM (VALUES (s.segment_start), (r.start_time), (p.process_start)) AS q(v)
) a
CROSS APPLY
(
    SELECT MIN(v) AS overlap_end
    FROM (VALUES (s.segment_end), (r.end_time), (p.process_end)) AS q(v)
) b
WHERE b.overlap_end > a.overlap_start
GROUP BY s.group_no;

-- Process Time "actual": irisan segment shift x proses SAJA (tanpa dibatasi runtime).
-- Dipakai HANYA untuk tampilan Process Time (mendekati aplikasi bawaan mesin).
-- Power On / Loss / Productivity tetap dari #process_totals (metode runtime-intersection).
DROP TABLE IF EXISTS #process_actual_totals;
CREATE TABLE #process_actual_totals
(
    group_no INT PRIMARY KEY,
    process_seconds BIGINT NOT NULL
);

INSERT INTO #process_actual_totals (group_no, process_seconds)
SELECT
    s.group_no,
    SUM(DATEDIFF(SECOND, a.overlap_start, b.overlap_end))
FROM #segments s
INNER JOIN #process_intervals p
    ON p.process_start < s.segment_end
   AND p.process_end > s.segment_start
   AND s.segment_start < @cutoff_time
CROSS APPLY
(
    SELECT MAX(v) AS overlap_start
    FROM (VALUES (s.segment_start), (p.process_start)) AS q(v)
) a
CROSS APPLY
(
    SELECT MIN(v) AS overlap_end
    FROM (VALUES (s.segment_end), (p.process_end)) AS q(v)
) b
WHERE b.overlap_end > a.overlap_start
GROUP BY s.group_no;

;WITH result AS
(
    SELECT
        g.group_no,
        g.group_name,
        g.period_start,
        g.period_end,
        g.break_start,
        g.break_end,
        ISNULL(pw.power_seconds, 0) AS power_seconds,
        ISNULL(pr.process_seconds, 0) AS process_seconds,
        ISNULL(pa.process_seconds, 0) AS process_actual_seconds
    FROM #groups g
    LEFT JOIN #power_totals pw ON pw.group_no = g.group_no
    LEFT JOIN #process_totals pr ON pr.group_no = g.group_no
    LEFT JOIN #process_actual_totals pa ON pa.group_no = g.group_no
),
final AS
(
    SELECT
        *,
        CASE WHEN power_seconds > process_seconds THEN power_seconds - process_seconds ELSE 0 END AS loss_seconds,
        CASE
            WHEN power_seconds > 0
            THEN CAST(process_seconds * 100.0 / power_seconds AS DECIMAL(7,2))
            ELSE CAST(0 AS DECIMAL(7,2))
        END AS productivity
    FROM result
)
SELECT
    UPPER(@mode) AS mode,
    CONVERT(varchar(10), @work_date, 23) AS work_date,
    CASE WHEN UPPER(@mode) = N'SHIFT' THEN @area ELSE NULL END AS area,
    group_name AS shift_name,
    CONVERT(varchar(19), period_start, 120) AS shift_start,
    CONVERT(varchar(19), period_end, 120) AS shift_end,
    CONVERT(varchar(19), break_start, 120) AS break_start,
    CONVERT(varchar(19), break_end, 120) AS break_end,
    power_seconds,
    process_seconds,
    loss_seconds,
    productivity,
    process_actual_seconds
FROM final
ORDER BY group_no;
`
