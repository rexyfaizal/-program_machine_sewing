<script setup>
import { computed, ref } from "vue";
import {
  MONTH_LABELS_ID,
  formatDateRangeLabel,
  getMonthRangeByIndex,
  getMonthWeekOptions,
  todayLocal,
} from "../../utils/format";

const props = defineProps({
  loading: {
    type: Boolean,
    default: false,
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
  locationValue: {
    type: String,
    default: "ALL",
  },
  locationOptions: {
    type: Array,
    default: () => [],
  },
  shiftValue: {
    type: String,
    default: "CURRENT",
  },
  shiftOptions: {
    type: Array,
    default: () => [],
  },
  showShiftFilter: {
    type: Boolean,
    default: false,
  },
  statusValue: {
    type: String,
    default: "ALL",
  },
  statusOptions: {
    type: Array,
    default: () => [
      { value: "ALL", label: "All Status" },
      { value: "GOOD", label: "GOOD" },
      { value: "NORMAL", label: "NORMAL" },
      { value: "BAD", label: "BAD" },
    ],
  },
});

const emit = defineEmits([
  "update:selectedDate",
  "update:startDate",
  "update:endDate",
  "update:keyword",
  "update:locationValue",
  "update:shiftValue",
  "update:statusValue",
  "dateChange",
  "download",
]);

const now = new Date();
const viewMode = ref("day");
const manualDate = ref(props.startDate || props.selectedDate || todayLocal());
const pickerYear = ref(now.getFullYear());
const pickerMonth = ref(now.getMonth());
const pickerWeek = ref(1);

const searchPlaceholder = computed(() => {
  if (props.showActions) {
    return "Search machines, UUID, status, area, location, operator, operator note...";
  }

  return "Search machines, status, area, location, operator logged in, operator note...";
});

const weekOptions = computed(() =>
  getMonthWeekOptions(pickerYear.value, pickerMonth.value)
);

function applyRange(start, end) {
  emit("update:startDate", start);
  emit("update:endDate", end);
  emit("update:selectedDate", start);
  emit("dateChange");
}

function applyCurrentSelection() {
  if (viewMode.value === "day") {
    const day = String(manualDate.value || todayLocal()).trim();
    applyRange(day, day);
    return;
  }

  if (viewMode.value === "week") {
    const weeks = weekOptions.value;
    const week =
      weeks.find((item) => item.weekNo === Number(pickerWeek.value)) || weeks[0];
    if (!week) return;
    pickerWeek.value = week.weekNo;
    applyRange(week.start, week.end);
    return;
  }

  const range = getMonthRangeByIndex(pickerYear.value, pickerMonth.value);
  applyRange(range.start, range.end);
}

function syncPickerFromManualDate() {
  const raw = String(manualDate.value || "").trim();
  const base = raw ? new Date(`${raw}T00:00:00`) : new Date();
  if (Number.isNaN(base.getTime())) return;
  pickerYear.value = base.getFullYear();
  pickerMonth.value = base.getMonth();
}

function updateViewMode(mode) {
  const next = String(mode || "day");
  if (viewMode.value === next) return;
  viewMode.value = next;
  syncPickerFromManualDate();
  if (viewMode.value === "week" && !weekOptions.value.some((item) => item.weekNo === pickerWeek.value)) {
    pickerWeek.value = weekOptions.value[0]?.weekNo || 1;
  }
  applyCurrentSelection();
}

function updateMonth(event) {
  pickerMonth.value = Number(event.target.value);
  if (viewMode.value === "week" && !weekOptions.value.some((item) => item.weekNo === pickerWeek.value)) {
    pickerWeek.value = weekOptions.value[0]?.weekNo || 1;
  }
  applyCurrentSelection();
}

function updateWeek(event) {
  pickerWeek.value = Number(event.target.value);
  applyCurrentSelection();
}

function updateManualDate(event) {
  const value = String(event.target.value || "").trim();
  if (!value) return;
  manualDate.value = value;
  syncPickerFromManualDate();
  if (viewMode.value === "week" && !weekOptions.value.some((item) => item.weekNo === pickerWeek.value)) {
    pickerWeek.value = weekOptions.value[0]?.weekNo || 1;
  }
  applyCurrentSelection();
}

function updateKeyword(event) {
  emit("update:keyword", event.target.value);
}

function updateLocation(event) {
  emit("update:locationValue", event.target.value);
  emit("dateChange");
}

function updateShift(event) {
  emit("update:shiftValue", event.target.value);
  emit("dateChange");
}

function updateStatus(event) {
  emit("update:statusValue", event.target.value);
}

const rangeLabel = computed(() => {
  const start = String(props.startDate || props.selectedDate || manualDate.value || "").trim();
  const end = String(props.endDate || start).trim();
  return formatDateRangeLabel(start, end);
});
</script>

<template>
  <div class="table-toolbar simple-toolbar">
    <label class="table-filter date-filter">
      <span>Tanggal</span>
      <input
        type="date"
        :value="manualDate"
        @input="updateManualDate"
      />
    </label>

    <div class="table-filter view-filter">
      <span>Periode</span>
      <div class="view-switch">
        <button
          type="button"
          class="view-btn"
          :class="{ active: viewMode === 'day' }"
          @click="updateViewMode('day')"
        >
          Hari
        </button>
        <button
          type="button"
          class="view-btn"
          :class="{ active: viewMode === 'week' }"
          @click="updateViewMode('week')"
        >
          Minggu
        </button>
        <button
          type="button"
          class="view-btn"
          :class="{ active: viewMode === 'month' }"
          @click="updateViewMode('month')"
        >
          Bulan
        </button>
      </div>
    </div>

    <label v-if="viewMode === 'week'" class="table-filter">
      <span>Pilih minggu</span>
      <select :value="pickerWeek" @change="updateWeek">
        <option
          v-for="week in weekOptions"
          :key="week.weekNo"
          :value="week.weekNo"
        >
          Minggu {{ week.weekNo }} ({{ week.rangeLabel }})
        </option>
      </select>
    </label>

    <label v-if="viewMode === 'month'" class="table-filter">
      <span>Pilih bulan</span>
      <select :value="pickerMonth" @change="updateMonth">
        <option
          v-for="(name, index) in MONTH_LABELS_ID"
          :key="name"
          :value="index"
        >
          {{ name }}
        </option>
      </select>
    </label>

    <label class="table-filter area-filter">
      <span>Area</span>
      <select :value="props.locationValue" @change="updateLocation">
        <option value="ALL">All GM</option>
        <option
          v-for="location in props.locationOptions"
          :key="location"
          :value="location"
        >
          {{ location }}
        </option>
      </select>
    </label>

    <label v-if="props.showShiftFilter" class="table-filter area-filter">
      <span>Shift</span>
      <select :value="props.shiftValue" @change="updateShift">
        <option
          v-for="shift in props.shiftOptions"
          :key="shift.value"
          :value="shift.value"
        >
          {{ shift.label }}
        </option>
      </select>
    </label>

    <label class="table-filter area-filter">
      <span>Status</span>
      <select :value="props.statusValue" @change="updateStatus">
        <option
          v-for="status in props.statusOptions"
          :key="status.value"
          :value="status.value"
        >
          {{ status.label }}
        </option>
      </select>
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
      :disabled="props.loading"
      @click="emit('download')"
    >
      Export Excel
    </button>
  </div>
  <p class="range-line">Range: {{ rangeLabel }}</p>
</template>
