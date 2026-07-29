<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
  machines: {
    type: Array,
    required: true,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
  pageSize: {
    type: Number,
    default: 10,
  },
  showActions: {
    type: Boolean,
    default: false,
  },
  selectedDate: {
    type: String,
    default: "",
  },
  startDate: {
    type: String,
    default: "",
  },
  endDate: {
    type: String,
    default: "",
  },
  keyword: {
    type: String,
    default: "",
  },
});

const emit = defineEmits([
  "edit",
  "update:selectedDate",
  "update:startDate",
  "update:endDate",
  "update:keyword",
  "dateChange",
  "download",
]);

const currentPage = ref(1);

const totalRows = computed(() => props.machines.length);

const totalPages = computed(() => {
  return Math.max(1, Math.ceil(totalRows.value / props.pageSize));
});

const startIndex = computed(() => {
  return (currentPage.value - 1) * props.pageSize;
});

const endIndex = computed(() => {
  return Math.min(startIndex.value + props.pageSize, totalRows.value);
});

const sortedMachines = computed(() => {
  return [...props.machines].sort((a, b) => {
    return Number(b.productivity || 0) - Number(a.productivity || 0);
  });
});

const pagedMachines = computed(() => {
  return sortedMachines.value.slice(startIndex.value, endIndex.value);
});

const totalCols = computed(() => {
  return props.showActions ? 18 : 10;
});

const searchPlaceholder = computed(() => {
  if (props.showActions) {
    return "Search machines, UUID, status, location, operator, operator note...";
  }

  return "Search machines, status, location, operator logged in, operator note...";
});

watch(
  () => totalRows.value,
  () => {
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value;
    }

    if (currentPage.value < 1) {
      currentPage.value = 1;
    }
  }
);

watch(
  () => props.pageSize,
  () => {
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value;
    }
  }
);

function toNumber(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function downTimeSec(machine) {
  const powerOnDuration = toNumber(machine?.runtime);
  const runningTime = toNumber(machine?.procTime);

  return Math.max(0, powerOnDuration - runningTime);
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--;
  }
}

function goToPage(page) {
  currentPage.value = page;
}

function editMachine(machine) {
  emit("edit", machine);
}

function statusClass(status) {
  const s = String(status || "BAD").toUpperCase();

  if (s === "GOOD") return "good";
  if (s === "NORMAL") return "normal";
  return "bad";
}

function formatHour(seconds) {
  const totalSeconds = Math.max(0, Math.floor(toNumber(seconds)));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);

  return `${hours}h ${minutes}m`;
}

function getOperatorNote(machine) {
  return machine?.operatorNote || machine?.spv || "";
}

function getOperatorName(machine) {
  return machine?.pic || "";
}

function getOperatorSubText(machine) {
  return machine?.operatorSubText || "";
}

function updateStartDate(event) {
  const value = event.target.value;

  emit("update:startDate", value);
  emit("update:selectedDate", value);
}

function updateEndDate(event) {
  emit("update:endDate", event.target.value);
}

function updateKeyword(event) {
  emit("update:keyword", event.target.value);
}

const visiblePages = computed(() => {
  const pages = [];
  const total = totalPages.value;
  const current = currentPage.value;

  let start = Math.max(1, current - 2);
  let end = Math.min(total, current + 2);

  if (current <= 3) {
    end = Math.min(total, 5);
  }

  if (current >= total - 2) {
    start = Math.max(1, total - 4);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});
</script>

<template>
  <section class="panel table-panel">
    <div class="panel-head modern-head">
      <div>
        <h2>Machine Productivity Details</h2>
      </div>

      <span>{{ totalRows }} data</span>
    </div>

    <div class="table-toolbar">
      <label class="table-filter date-filter">
        <span>Start Date</span>

        <input
          type="date"
          :value="props.startDate || props.selectedDate"
          @input="updateStartDate"
          @change="emit('dateChange')"
        />
      </label>

      <label class="table-filter date-filter">
        <span>End Date</span>

        <input
          type="date"
          :value="props.endDate || props.startDate || props.selectedDate"
          @input="updateEndDate"
        />
      </label>

      <label class="table-filter search-filter">
        <span>Search</span>

        <input
          type="text"
          :value="props.keyword"
          :placeholder="searchPlaceholder"
          @input="updateKeyword"
        />
      </label>

      <button
        type="button"
        class="download-btn"
        :disabled="loading"
        @click="emit('download')"
      >
        Export Excel
      </button>
    </div>

    <div class="table-wrap">
      <table :class="{ 'user-table': !showActions }">
        <thead>
          <tr v-if="showActions">
            <th>No</th>
            <th class="left">Machine Name</th>
            <th>IP</th>
            <th>Location</th>
            <th>Operator</th>
            <th>Operator Note</th>
            <th class="center">Produktivitas</th>
            <th>Status</th>
            <th class="center">Power On Duration</th>
            <th class="center">Running Time</th>
            <th class="center">Loss Time</th>
            <th>Output</th>
            <th>Abnormal</th>
            <th>Avg / Max CT</th>
            <th>Alarm</th>
            <th>Program Dominan</th>
            <th>UUID / Table</th>
            <th class="center">Action</th>
          </tr>

          <tr v-else>
            <th>No</th>
            <th class="left">Machine Name</th>
            <th>Location</th>
            <th>Operator</th>
            <th>Operator Note</th>
            <th class="center">Power On Duration</th>
            <th class="center">Running Time</th>
            <th class="center">Loss Time</th>
            <th class="center">Produktivitas</th>
            <th class="center">Status</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="(m, index) in pagedMachines" :key="`${m.uuid}-${index}`">
            <template v-if="showActions">
              <td class="right">
                {{ startIndex + index + 1 }}
              </td>

              <td>
                <strong>{{ m.machineName || "-" }}</strong>
                <small v-if="m.customName">
                  Original: {{ m.originalMachineName || "-" }}
                </small>
              </td>

              <td class="mono">
                {{ m.ip || "-" }}
              </td>

              <td>
                <strong>{{ m.location || "-" }}</strong>
              </td>

              <td class="operator-cell">
                <strong v-if="getOperatorName(m)" class="operator-name">
                  {{ getOperatorName(m) }}
                </strong>

                <strong v-else class="operator-empty">
                  Not logged in
                </strong>

                <small v-if="getOperatorSubText(m)" class="operator-sub">
                  {{ getOperatorSubText(m) }}
                </small>
              </td>

              <td class="note-cell">
                {{ getOperatorNote(m) || "-" }}
              </td>

              <td class="center productivity-cell">
                {{ Number(m.productivity || 0).toFixed(2) }}%
              </td>

              <td>
                <span class="badge" :class="statusClass(m.status)">
                  {{ m.status || "BAD" }}
                </span>
              </td>

              <td class="center time-cell">
                {{ formatHour(m.runtime) }}
              </td>

              <td class="center time-cell">
                {{ formatHour(m.procTime) }}
              </td>

              <td class="center time-cell">
                {{ formatHour(downTimeSec(m)) }}
              </td>

              <td class="right">
                <strong>{{ m.output || 0 }}</strong>
                <small>{{ m.complete || 0 }} complete</small>
              </td>

              <td class="right">
                {{ m.abnormal || 0 }}
              </td>

              <td class="right">
                {{ Number(m.avgCT || 0).toFixed(2) }}s
                <small>Max {{ Number(m.maxCT || 0).toFixed(0) }}s</small>
              </td>

              <td class="right">
                {{ m.alarm || 0 }}
                <small>{{ m.alarmTypes || "-" }}</small>
              </td>

              <td>
                {{ m.program || "-" }}
                <small>{{ m.firstProcess || "-" }} → {{ m.lastProcess || "-" }}</small>
              </td>

              <td class="mono">
                {{ m.uuid || "-" }}
                <small>{{ m.tableName || "-" }}</small>
              </td>

              <td class="center">
                <button class="btn-edit" type="button" @click="editMachine(m)">
                  Edit
                </button>
              </td>
            </template>

            <template v-else>
              <td class="right">
                {{ startIndex + index + 1 }}
              </td>

              <td>
                <strong>{{ m.machineName || "-" }}</strong>
              </td>

              <td>
                {{ m.location || "-" }}
              </td>

              <td class="operator-cell">
                <strong v-if="getOperatorName(m)" class="operator-name">
                  {{ getOperatorName(m) }}
                </strong>

                <strong v-else class="operator-empty">
                  Not logged in
                </strong>

                <small v-if="getOperatorSubText(m)" class="operator-sub">
                  {{ getOperatorSubText(m) }}
                </small>
              </td>

              <td class="note-cell">
                {{ getOperatorNote(m) || "-" }}
              </td>

              <td class="center time-cell">
                {{ formatHour(m.runtime) }}
              </td>

              <td class="center time-cell">
                {{ formatHour(m.procTime) }}
              </td>

              <td class="center time-cell">
                {{ formatHour(downTimeSec(m)) }}
              </td>

              <td class="center productivity-cell">
                {{ Number(m.productivity || 0).toFixed(2) }}%
              </td>

              <td class="center">
                <span class="badge" :class="statusClass(m.status)">
                  {{ m.status || "BAD" }}
                </span>
              </td>
            </template>
          </tr>

          <tr v-if="!pagedMachines.length && !loading">
            <td :colspan="totalCols" class="empty">
              Data tidak ditemukan.
            </td>
          </tr>

          <tr v-if="loading">
            <td :colspan="totalCols" class="empty">
              Mengambil data...
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination">
      <div class="page-info">
        <span v-if="totalRows">
          Menampilkan {{ startIndex + 1 }} - {{ endIndex }} dari {{ totalRows }} data
        </span>
        <span v-else>
          Tidak ada data
        </span>
      </div>

      <div class="page-controls">
        <button @click="prevPage" :disabled="currentPage <= 1">
          Prev
        </button>

        <button
          v-for="page in visiblePages"
          :key="page"
          class="page-number"
          :class="{ active: currentPage === page }"
          @click="goToPage(page)"
        >
          {{ page }}
        </button>

        <button @click="nextPage" :disabled="currentPage >= totalPages">
          Next
        </button>
      </div>

      <div class="page-total">
        Page {{ currentPage }} / {{ totalPages }}
      </div>
    </div>
  </section>
</template>

<style scoped>
.table-panel {
  margin-bottom: 30px;
  min-width: 0;
}

.panel {
  background: linear-gradient(135deg, #ffffff, #f8fbff);
  border: 1px solid #dbeafe;
  border-radius: 24px;
  padding: 20px;
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.05);
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 14px;
}

.panel-head h2 {
  margin: 0;
  font-size: 20px;
  letter-spacing: -0.02em;
  color: #0f172a;
}

.panel-head p {
  margin: 5px 0 0;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.panel-head span {
  color: #64748b;
  font-size: 13px;
  font-weight: 900;
  white-space: nowrap;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  padding: 7px 10px;
  border-radius: 999px;
}

.table-toolbar {
  display: grid;
  grid-template-columns:
    minmax(150px, 190px)
    minmax(150px, 190px)
    minmax(240px, 1fr)
    minmax(150px, 180px);
  gap: 10px;
  margin-bottom: 12px;
  align-items: stretch;
}

.table-filter {
  min-width: 0;
  display: grid;
  gap: 3px;
  padding: 8px 12px;
  border-radius: 13px;
  background: #ffffff;
  border: 1px solid #dbeafe;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.035);
}

.table-filter span {
  color: #64748b;
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.table-filter input {
  width: 100%;
  min-width: 0;
  height: 24px;
  border: 0;
  outline: none;
  background: transparent;
  color: #0f172a;
  font-size: 13px;
  font-weight: 800;
}

.table-filter input::placeholder {
  color: #94a3b8;
  font-weight: 700;
}

.download-btn {
  border: 0;
  border-radius: 13px;
  color: #ffffff;
  background: linear-gradient(135deg, #21a366, #107c41);
  box-shadow: 0 12px 24px rgba(16, 124, 65, 0.24);
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
  min-height: 50px;
  padding: 0 16px;
  white-space: nowrap;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    filter 0.2s ease,
    opacity 0.2s ease;
}

.download-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.03);
  box-shadow: 0 16px 28px rgba(16, 124, 65, 0.32);
}

.download-btn:active:not(:disabled) {
  transform: translateY(0);
}

.download-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
  filter: none;
}

.table-wrap {
  width: 100%;
  overflow-x: auto;
  border: 1px solid #dbe4ef;
  border-radius: 16px;
}

table {
  width: 100%;
  min-width: 1380px;
  border-collapse: collapse;
  background: white;
  font-size: 13px;
}

table.user-table {
  min-width: 1180px;
}

th {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  color: #ffffff;
  text-align: left;
  padding: 12px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #1e40af;
  white-space: nowrap;
}

th + th {
  border-left: 1px solid rgba(255, 255, 255, 0.18);
}

td {
  padding: 12px;
  border-bottom: 1px solid #edf2f7;
  vertical-align: top;
  color: #1e293b;
}

td strong {
  font-weight: 900;
}

td small {
  display: block;
  color: #64748b;
  font-size: 11px;
  margin-top: 4px;
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

tr:hover td {
  background: #fbfdff;
}

.right {
  text-align: right;
}

.center {
  text-align: center;
}

.center small {
  margin-left: auto;
  margin-right: auto;
  text-align: center;
}

.time-cell,
.productivity-cell {
  font-weight: 500;
  color: #1e293b;
  white-space: nowrap;
}

.operator-cell {
  min-width: 150px;
  max-width: 220px;
  white-space: normal;
  line-height: 1.35;
}

.operator-name {
  display: block;
  color: #0f172a;
  font-size: 12px;
  font-weight: 950;
  white-space: normal;
}

.operator-empty {
  display: block;
  color: #94a3b8;
  font-size: 12px;
  font-weight: 900;
}

.operator-sub {
  display: block;
  margin-top: 4px;
  color: #2563eb !important;
  font-size: 11px !important;
  font-weight: 900;
  max-width: none !important;
  overflow: visible !important;
  text-overflow: unset !important;
  white-space: normal !important;
}

.note-cell {
  min-width: 170px;
  max-width: 260px;
  white-space: normal;
  line-height: 1.35;
  font-size: 12px;
  font-weight: 800;
  color: #334155;
}

.mono {
  font-family: Consolas, Menlo, monospace;
  font-size: 12px;
}

.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: fit-content;
  border-radius: 999px;
  padding: 7px 10px;
  font-size: 12px;
  font-weight: 900;
}

.good {
  background: #dcfce7;
  color: #166534;
}

.normal {
  background: #fef3c7;
  color: #92400e;
}

.bad {
  background: #fee2e2;
  color: #991b1b;
}

.btn-edit {
  border: 1px solid #dbe4ef;
  background: #ffffff;
  color: #0f172a;
  border-radius: 10px;
  padding: 8px 12px;
  font-weight: 900;
  cursor: pointer;
}

.btn-edit:hover {
  background: #f1f5f9;
}

.empty {
  text-align: center;
  padding: 30px;
  color: #64748b;
  font-weight: 800;
}

.pagination {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 14px;
  align-items: center;
  margin-top: 14px;
  color: #64748b;
  font-size: 13px;
  font-weight: 800;
}

.page-controls {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.page-controls button {
  border: 1px solid #dbe4ef;
  background: white;
  color: #0f172a;
  border-radius: 10px;
  padding: 8px 12px;
  font-weight: 900;
  cursor: pointer;
}

.page-controls button:hover:not(:disabled) {
  background: #f1f5f9;
}

.page-controls button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.page-number.active {
  background: #2563eb;
  color: white;
  border-color: #2563eb;
}

.page-total {
  text-align: right;
}

@media (max-width: 1200px) {
  .table-toolbar {
    grid-template-columns: minmax(150px, 190px) minmax(150px, 190px) 1fr;
  }

  .download-btn {
    grid-column: 1 / -1;
  }
}

@media (max-width: 980px) {
  .table-toolbar {
    grid-template-columns: 1fr;
  }

  .download-btn {
    min-height: 50px;
  }
}

@media (max-width: 900px) {
  .pagination {
    grid-template-columns: 1fr;
    text-align: center;
  }

  .page-total {
    text-align: center;
  }

  .page-controls {
    flex-wrap: wrap;
  }
}
</style>