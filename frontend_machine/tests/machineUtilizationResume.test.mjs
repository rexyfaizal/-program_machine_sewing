import { after, before, test } from "node:test";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import { createServer } from "vite";

let server;
let resume;
const weeks = [
  { weekNo: 1, start: "2026-10-01", end: "2026-10-03", rangeLabel: "1-3 Oktober" },
  { weekNo: 2, start: "2026-10-05", end: "2026-10-10", rangeLabel: "5-10 Oktober" },
];
const machine = (overrides = {}) => ({ uuid: "machine-A", location: "GM1 LINE 2", nickName: "Jack", runtimeSec: 0, procSec: 0, ...overrides });
const session = (overrides = {}) => ({ ...machine(), id: 1, operatorNik: "000123", operatorName: "ANI", processName: "Jahit", hasSessionStats: true, ...overrides });
const day = (date, machines, operators = []) => ({ date, productivity: { rows: machines }, operators });

before(async () => {
  server = await createServer({ configFile: false,
    root: fileURLToPath(new URL("..", import.meta.url)),
    server: { middlewareMode: true }, appType: "custom" });
  resume = await server.ssrLoadModule("/src/utils/machineUtilizationResume.js");
});
after(async () => { await server?.close(); });

test("combines sessions by day before averaging daily utilization, ignoring duplicate session IDs", () => {
  const first = session({ runtimeSec: 3600, procSec: 1800 });
  const result = resume.buildMachineUtilizationResume([
    day("2026-10-01", [machine()], [first, first, session({ id: 2, runtimeSec: 3600, procSec: 3600 })]),
    day("2026-10-02", [machine()], [session({ runtimeSec: 36000, procSec: 9000 })]),
  ], weeks);
  assert.equal(result.length, 1);
  assert.equal(result[0].weeks[1].value, 50); // Average of 75% and 25%, not duration-weighted.
  assert.equal(result[0].weeks[1].days, 2);
  assert.equal(result[0].weeks[2].value, null);
});

test("distinguishes measured zero utilization from an unmeasured day or week", () => {
  const [row] = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine()], [session({ runtimeSec: 3600, procSec: 0 })]),
    day("2026-10-06", [machine()], [session()]),
  ], weeks);
  assert.equal(row.weeks[1].value, null);
  assert.equal(row.weeks[2].value, 0);
  assert.equal(row.weeks[2].days, 1);
});

test("uses each operator's session metrics and retains automatic logout sessions", () => {
  const result = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine({ runtimeSec: 10000, procSec: 1000 })], [
      session({ runtimeSec: 100, procSec: 90, status: "AUTO_LOGOUT_WORK_END" }),
      session({ id: 2, operatorNik: "000456", operatorName: "BUDI", runtimeSec: 100, procSec: 10 }),
    ]),
  ], weeks);
  assert.equal(result.length, 2);
  assert.equal(result.find((row) => row.operatorNik === "000123").weeks[2].value, 90);
  assert.equal(result.find((row) => row.operatorNik === "000456").weeks[2].value, 10);
});

test("preserves historical location and separates processes", () => {
  const result = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine({ location: "GM3 LINE 1" })], [
      session(),
      session({ id: 2, processName: "Potong" }),
    ]),
  ], weeks);
  assert.equal(result.length, 2);
  assert.ok(result.every((row) => row.location === "GM1 LINE 2" && row.area === "GM1"));
});

test("uses machine metrics for unassigned machines", () => {
  const [row] = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine({ runtimeSec: 100, procSec: 50 })]),
  ], weeks);
  assert.equal(row.operatorName, "");
  assert.equal(row.weeks[2].value, 50);
});

test("retains inactive machines with empty cells without adding phantom operator rows", () => {
  const result = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine(), machine({ uuid: "inactive" })], [session({ runtimeSec: 100, procSec: 50 })]),
    day("2026-10-06", [machine(), machine({ uuid: "inactive" })]),
  ], weeks);
  assert.equal(result.length, 2);
  assert.equal(result.filter((row) => row.uuid === "machine-A").length, 1);
  assert.equal(result.find((row) => row.uuid === "inactive").weeks[2].value, null);
});

test("applies the existing process/runtime tolerance consistently", () => {
  const [row] = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine({ runtimeSec: 200, procSec: 300 })]),
  ], weeks);
  assert.equal(row.weeks[2].value, 50); // Runtime becomes process + 300s.
});

test("building and keyword filters determine the rows and averages", () => {
  const rows = resume.buildMachineUtilizationResume([
    day("2026-10-05", [
      machine({ runtimeSec: 100, procSec: 50 }),
      machine({ uuid: "GM2-machine", location: "GM2 LINE 1", runtimeSec: 100, procSec: 0 }),
    ]),
  ], weeks);
  const filtered = resume.filterMachineUtilizationRows(rows, "GM1", "jack");
  assert.equal(filtered.length, 1);
  assert.equal(resume.weeklyAverages(filtered, weeks)[2], 50);
  assert.equal(resume.weeklyAverages(rows, weeks)[2], 25);
  assert.equal(resume.filterMachineUtilizationRows(rows, "GM3").length, 0);
  assert.equal(resume.average([null, undefined, NaN, 0, 50]), 25);
});

test("sorts line numbers naturally", () => {
  const rows = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine({ uuid: "ten", location: "GM1 LINE 10" }), machine({ uuid: "two", location: "GM1 LINE 2" })]),
  ], weeks);
  assert.deepEqual(rows.map((row) => row.uuid), ["two", "ten"]);
});

test("Excel preserves NIK strings, numeric percentages, selected weeks, blanks and average formulas", () => {
  const rows = resume.buildMachineUtilizationResume([
    day("2026-10-05", [machine()], [session({ runtimeSec: 100, procSec: 50 })]),
  ], weeks);
  const workbook = resume.createMachineUtilizationWorkbook({ rows, weeks, monthLabel: "Oktober 2026" });
  const sheet = workbook.Sheets[workbook.SheetNames[0]];
  assert.equal(sheet.E2.t, "s");
  assert.equal(sheet.E2.v, "000123");
  assert.equal(sheet.G2, undefined); // Missing W1 remains empty.
  assert.equal(sheet.H2.v, 0.5);
  assert.equal(sheet.H2.z, "0.00%");
  assert.equal(sheet.H3.v, 0.5);
  assert.equal(sheet.H3.f, "AVERAGE(H2:H2)");
  assert.equal(sheet.F3.v, "Average");
  assert.match(sheet.H1.v, /AVG W2\nOktober 2026\n5-10 Oktober/);
  assert.equal(sheet["!autofilter"].ref, "A1:H2");
  const selected = resume.createMachineUtilizationWorkbook({ rows, weeks: [weeks[1]], monthLabel: "Oktober 2026" });
  const selectedSheet = selected.Sheets[selected.SheetNames[0]];
  assert.equal(selectedSheet.G2.v, 0.5);
  assert.equal(selectedSheet.H1, undefined);
});
