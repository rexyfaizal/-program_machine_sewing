<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";
import {
  downTimeSec,
  formatHour,
  formatMachineIdLines,
  formatMinute,
  getDisplayMachineName,
  getOperatorDisplayRows,
  getOperatorNoteItems,
  isUsingProcessName,
  statusClass,
} from "../../utils/machineTableFormat";
import { parseLocationParts } from "../../utils/gm3Shift";

function getBranchLabel(machine) {
  const { factory } = parseLocationParts(machine?.location);
  return factory || "-";
}

function getLocationLineLabel(machine) {
  const raw = String(machine?.location || "").trim();
  if (!raw || raw === "-") return "-";
  const { lineName } = parseLocationParts(raw);
  return lineName || raw;
}

function getProsesIdFromOperator(op, machine) {
  return String(op?.prosesId || machine?.operatorProsesId || "").trim() || "-";
}

const props = defineProps({
  pagedMachines: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
  showActions: {
    type: Boolean,
    default: false,
  },
  startIndex: {
    type: Number,
    default: 0,
  },
  totalCols: {
    type: Number,
    default: 10,
  },
  sortKey: {
    type: String,
    default: "productivity",
  },
  sortDir: {
    type: String,
    default: "desc",
  },
  columnVisibility: {
    type: Object,
    default: () => ({}),
  },
});

function showCol(key) {
  if (key === "no") return true;
  return props.columnVisibility?.[key] !== false;
}

const emit = defineEmits(["edit", "chart", "ask-ai", "why-bad", "set-sort"]);

function onSort(key, dir) {
  emit("set-sort", key, dir);
}

const ctxMenu = ref(null);

function editMachine(machine) {
  emit("edit", machine);
}

function chartMachine(machine) {
  emit("chart", machine);
}

function askAiMachine(machine) {
  closeCtxMenu();
  emit("ask-ai", machine);
}

function whyBadMachine(machine) {
  closeCtxMenu();
  emit("why-bad", machine);
}

function closeCtxMenu() {
  ctxMenu.value = null;
}

function onRowContextMenu(event, machine) {
  if (!props.showActions) return;
  event.preventDefault();
  ctxMenu.value = {
    x: event.clientX,
    y: event.clientY,
    machine,
  };
}

function onGlobalClick() {
  closeCtxMenu();
}

onMounted(() => {
  window.addEventListener("click", onGlobalClick);
  window.addEventListener("scroll", onGlobalClick, true);
});

onBeforeUnmount(() => {
  window.removeEventListener("click", onGlobalClick);
  window.removeEventListener("scroll", onGlobalClick, true);
});
</script>

<template>
  <div class="table-wrap">
    <table :class="{ 'user-table': !props.showActions }">
      <thead>
        <tr v-if="props.showActions">
          <th v-if="showCol('no')">
            <button type="button" class="th-sort" @click="onSort('no')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'no' && sortDir === 'asc' }"
                  @click.stop="onSort('no', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'no' && sortDir === 'desc' }"
                  @click.stop="onSort('no', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">No</span>
            </button>
          </th>
          <th v-if="showCol('prosesId')" class="col-proses-id">
            <button type="button" class="th-sort" @click="onSort('prosesId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'prosesId' && sortDir === 'asc' }"
                  @click.stop="onSort('prosesId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'prosesId' && sortDir === 'desc' }"
                  @click.stop="onSort('prosesId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Proses ID</span>
            </button>
          </th>
          <th v-if="showCol('processName')" class="left">
            <button type="button" class="th-sort" @click="onSort('processName')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'processName' && sortDir === 'asc' }"
                  @click.stop="onSort('processName', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'processName' && sortDir === 'desc' }"
                  @click.stop="onSort('processName', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Process Name</span>
            </button>
          </th>
          <th v-if="showCol('style')" class="col-style">
            <button type="button" class="th-sort" @click="onSort('style')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'style' && sortDir === 'asc' }"
                  @click.stop="onSort('style', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'style' && sortDir === 'desc' }"
                  @click.stop="onSort('style', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Style</span>
            </button>
          </th>
          <th v-if="showCol('branch')">
            <button type="button" class="th-sort" @click="onSort('branch')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'branch' && sortDir === 'asc' }"
                  @click.stop="onSort('branch', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'branch' && sortDir === 'desc' }"
                  @click.stop="onSort('branch', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Branch</span>
            </button>
          </th>
          <th v-if="showCol('location')">
            <button type="button" class="th-sort" @click="onSort('location')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'location' && sortDir === 'asc' }"
                  @click.stop="onSort('location', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'location' && sortDir === 'desc' }"
                  @click.stop="onSort('location', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Line</span>
            </button>
          </th>
          <th v-if="showCol('spvId')" class="col-spv-id">
            <button type="button" class="th-sort" @click="onSort('spvId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'spvId' && sortDir === 'asc' }"
                  @click.stop="onSort('spvId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'spvId' && sortDir === 'desc' }"
                  @click.stop="onSort('spvId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">SPV_ID</span>
            </button>
          </th>
          <th v-if="showCol('machineId')" class="col-machine-id">
            <button type="button" class="th-sort" @click="onSort('machineId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'machineId' && sortDir === 'asc' }"
                  @click.stop="onSort('machineId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'machineId' && sortDir === 'desc' }"
                  @click.stop="onSort('machineId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Machine ID</span>
            </button>
          </th>
          <th v-if="showCol('mikaId')" class="col-mika-id">
            <button type="button" class="th-sort" @click="onSort('mikaId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'mikaId' && sortDir === 'asc' }"
                  @click.stop="onSort('mikaId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'mikaId' && sortDir === 'desc' }"
                  @click.stop="onSort('mikaId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Mika ID</span>
            </button>
          </th>
          <th v-if="showCol('nik')" class="col-nik">
            <button type="button" class="th-sort" @click="onSort('nik')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'nik' && sortDir === 'asc' }"
                  @click.stop="onSort('nik', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'nik' && sortDir === 'desc' }"
                  @click.stop="onSort('nik', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Operator_ID</span>
            </button>
          </th>
          <th v-if="showCol('namaOperator')">
            <button type="button" class="th-sort" @click="onSort('namaOperator')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'namaOperator' && sortDir === 'asc' }"
                  @click.stop="onSort('namaOperator', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'namaOperator' && sortDir === 'desc' }"
                  @click.stop="onSort('namaOperator', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Nama Operator</span>
            </button>
          </th>
          <th v-if="showCol('operatorNote')">
            <button type="button" class="th-sort" @click="onSort('operatorNote')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'operatorNote' && sortDir === 'asc' }"
                  @click.stop="onSort('operatorNote', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'operatorNote' && sortDir === 'desc' }"
                  @click.stop="onSort('operatorNote', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Operator Note</span>
            </button>
          </th>
          <th v-if="showCol('productivity')" class="center">
            <button type="button" class="th-sort" @click="onSort('productivity')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'productivity' && sortDir === 'asc' }"
                  @click.stop="onSort('productivity', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'productivity' && sortDir === 'desc' }"
                  @click.stop="onSort('productivity', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Utilisasi Mesin</span>
            </button>
          </th>
          <th v-if="showCol('status')">
            <button type="button" class="th-sort" @click="onSort('status')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'status' && sortDir === 'asc' }"
                  @click.stop="onSort('status', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'status' && sortDir === 'desc' }"
                  @click.stop="onSort('status', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Status</span>
            </button>
          </th>
          <th v-if="showCol('runtime')" class="center">
            <button type="button" class="th-sort" @click="onSort('runtime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'runtime' && sortDir === 'asc' }"
                  @click.stop="onSort('runtime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'runtime' && sortDir === 'desc' }"
                  @click.stop="onSort('runtime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Power On Duration
                <small class="th-sub">mesin nyala</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('procTime')" class="center">
            <button type="button" class="th-sort" @click="onSort('procTime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'procTime' && sortDir === 'asc' }"
                  @click.stop="onSort('procTime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'procTime' && sortDir === 'desc' }"
                  @click.stop="onSort('procTime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Running Time
                <small class="th-sub">mesin jalan</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('lossTime')" class="center">
            <button type="button" class="th-sort" @click="onSort('lossTime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'lossTime' && sortDir === 'asc' }"
                  @click.stop="onSort('lossTime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'lossTime' && sortDir === 'desc' }"
                  @click.stop="onSort('lossTime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Loss Time
                <small class="th-sub">waktu terbuang</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('output')">
            <button type="button" class="th-sort" @click="onSort('output')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'output' && sortDir === 'asc' }"
                  @click.stop="onSort('output', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'output' && sortDir === 'desc' }"
                  @click.stop="onSort('output', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Output</span>
            </button>
          </th>
          <th v-if="showCol('uuid')">
            <button type="button" class="th-sort" @click="onSort('uuid')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'uuid' && sortDir === 'asc' }"
                  @click.stop="onSort('uuid', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'uuid' && sortDir === 'desc' }"
                  @click.stop="onSort('uuid', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">UUID / Table</span>
            </button>
          </th>
          <th v-if="showCol('action')" class="center">Action</th>
          <th v-if="showCol('ai')" class="center">AI</th>
          <th v-if="showCol('grafik')" class="center">Grafik</th>
        </tr>

        <tr v-else>
          <th v-if="showCol('no')">
            <button type="button" class="th-sort" @click="onSort('no')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'no' && sortDir === 'asc' }"
                  @click.stop="onSort('no', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'no' && sortDir === 'desc' }"
                  @click.stop="onSort('no', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">No</span>
            </button>
          </th>
          <th v-if="showCol('prosesId')" class="col-proses-id">
            <button type="button" class="th-sort" @click="onSort('prosesId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'prosesId' && sortDir === 'asc' }"
                  @click.stop="onSort('prosesId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'prosesId' && sortDir === 'desc' }"
                  @click.stop="onSort('prosesId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Proses ID</span>
            </button>
          </th>
          <th v-if="showCol('processName')" class="left">
            <button type="button" class="th-sort" @click="onSort('processName')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'processName' && sortDir === 'asc' }"
                  @click.stop="onSort('processName', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'processName' && sortDir === 'desc' }"
                  @click.stop="onSort('processName', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Process Name</span>
            </button>
          </th>
          <th v-if="showCol('style')" class="col-style">
            <button type="button" class="th-sort" @click="onSort('style')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'style' && sortDir === 'asc' }"
                  @click.stop="onSort('style', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'style' && sortDir === 'desc' }"
                  @click.stop="onSort('style', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Style</span>
            </button>
          </th>
          <th v-if="showCol('branch')">
            <button type="button" class="th-sort" @click="onSort('branch')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'branch' && sortDir === 'asc' }"
                  @click.stop="onSort('branch', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'branch' && sortDir === 'desc' }"
                  @click.stop="onSort('branch', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Branch</span>
            </button>
          </th>
          <th v-if="showCol('location')">
            <button type="button" class="th-sort" @click="onSort('location')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'location' && sortDir === 'asc' }"
                  @click.stop="onSort('location', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'location' && sortDir === 'desc' }"
                  @click.stop="onSort('location', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Line</span>
            </button>
          </th>
          <th v-if="showCol('spvId')" class="col-spv-id">
            <button type="button" class="th-sort" @click="onSort('spvId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'spvId' && sortDir === 'asc' }"
                  @click.stop="onSort('spvId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'spvId' && sortDir === 'desc' }"
                  @click.stop="onSort('spvId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">SPV_ID</span>
            </button>
          </th>
          <th v-if="showCol('machineId')" class="col-machine-id">
            <button type="button" class="th-sort" @click="onSort('machineId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'machineId' && sortDir === 'asc' }"
                  @click.stop="onSort('machineId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'machineId' && sortDir === 'desc' }"
                  @click.stop="onSort('machineId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Machine ID</span>
            </button>
          </th>
          <th v-if="showCol('mikaId')" class="col-mika-id">
            <button type="button" class="th-sort" @click="onSort('mikaId')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'mikaId' && sortDir === 'asc' }"
                  @click.stop="onSort('mikaId', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'mikaId' && sortDir === 'desc' }"
                  @click.stop="onSort('mikaId', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Mika ID</span>
            </button>
          </th>
          <th v-if="showCol('nik')" class="col-nik">
            <button type="button" class="th-sort" @click="onSort('nik')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'nik' && sortDir === 'asc' }"
                  @click.stop="onSort('nik', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'nik' && sortDir === 'desc' }"
                  @click.stop="onSort('nik', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Operator_ID</span>
            </button>
          </th>
          <th v-if="showCol('namaOperator')">
            <button type="button" class="th-sort" @click="onSort('namaOperator')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'namaOperator' && sortDir === 'asc' }"
                  @click.stop="onSort('namaOperator', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'namaOperator' && sortDir === 'desc' }"
                  @click.stop="onSort('namaOperator', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Nama Operator</span>
            </button>
          </th>
          <th v-if="showCol('operatorNote')">
            <button type="button" class="th-sort" @click="onSort('operatorNote')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'operatorNote' && sortDir === 'asc' }"
                  @click.stop="onSort('operatorNote', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'operatorNote' && sortDir === 'desc' }"
                  @click.stop="onSort('operatorNote', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Operator Note</span>
            </button>
          </th>
          <th v-if="showCol('runtime')" class="center">
            <button type="button" class="th-sort" @click="onSort('runtime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'runtime' && sortDir === 'asc' }"
                  @click.stop="onSort('runtime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'runtime' && sortDir === 'desc' }"
                  @click.stop="onSort('runtime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Power On Duration
                <small class="th-sub">mesin nyala</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('procTime')" class="center">
            <button type="button" class="th-sort" @click="onSort('procTime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'procTime' && sortDir === 'asc' }"
                  @click.stop="onSort('procTime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'procTime' && sortDir === 'desc' }"
                  @click.stop="onSort('procTime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Running Time
                <small class="th-sub">mesin jalan</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('lossTime')" class="center">
            <button type="button" class="th-sort" @click="onSort('lossTime')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'lossTime' && sortDir === 'asc' }"
                  @click.stop="onSort('lossTime', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'lossTime' && sortDir === 'desc' }"
                  @click.stop="onSort('lossTime', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">
                Loss Time
                <small class="th-sub">waktu terbuang</small>
              </span>
            </button>
          </th>
          <th v-if="showCol('productivity')" class="center">
            <button type="button" class="th-sort" @click="onSort('productivity')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'productivity' && sortDir === 'asc' }"
                  @click.stop="onSort('productivity', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'productivity' && sortDir === 'desc' }"
                  @click.stop="onSort('productivity', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Utilisasi Mesin</span>
            </button>
          </th>
          <th v-if="showCol('status')" class="center">
            <button type="button" class="th-sort" @click="onSort('status')">
              <span class="th-sort__icons" aria-hidden="true">
                <span
                  class="th-sort__asc"
                  :class="{ active: sortKey === 'status' && sortDir === 'asc' }"
                  @click.stop="onSort('status', 'asc')"
                >▲</span>
                <span
                  class="th-sort__desc"
                  :class="{ active: sortKey === 'status' && sortDir === 'desc' }"
                  @click.stop="onSort('status', 'desc')"
                >▼</span>
              </span>
              <span class="th-sort__label">Status</span>
            </button>
          </th>
          <th v-if="showCol('grafik')" class="center">Grafik</th>
        </tr>
      </thead>

      <tbody>
        <tr
          v-for="(m, index) in props.pagedMachines"
          :key="`${m.uuid}-${index}`"
          @contextmenu="onRowContextMenu($event, m)"
        >
          <template v-if="props.showActions">
            <td v-if="showCol('no')" class="right">
              {{ m.__rowNo || props.startIndex + index + 1 }}
            </td>

            <td v-if="showCol('prosesId')" class="col-proses-id">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-proses-id-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">{{ getProsesIdFromOperator(op, m) }}</strong>
                </div>
              </template>
              <strong v-else class="operator-name">{{ m.operatorProsesId || "-" }}</strong>
            </td>

            <td v-if="showCol('processName')">
              <strong>{{ getDisplayMachineName(m) }}</strong>

              <small v-if="!isUsingProcessName(m) && m.customName">
                Original: {{ m.originalMachineName || "-" }}
              </small>
            </td>

            <td v-if="showCol('style')" class="col-style">
              {{ m.operatorStyleName || "-" }}
            </td>

            <td v-if="showCol('branch')">
              <strong>{{ getBranchLabel(m) }}</strong>
            </td>

            <td v-if="showCol('location')">
              <strong>{{ getLocationLineLabel(m) }}</strong>
            </td>

            <td v-if="showCol('spvId')" class="col-spv-id" :title="m.spvName || ''">
              <strong>{{ m.spvId || "-" }}</strong>
            </td>

            <td v-if="showCol('machineId')" class="col-machine-id" :title="m.machineId || ''">
              <div class="machine-id-stack">
                <span
                  v-for="(line, lineIndex) in formatMachineIdLines(m.machineId)"
                  :key="`${m.uuid || index}-mid-${lineIndex}`"
                >
                  {{ line }}
                </span>
              </div>
            </td>

            <td v-if="showCol('mikaId')" class="col-mika-id">-</td>

            <td v-if="showCol('nik')" class="col-nik">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-nik-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">{{ op.nik || "-" }}</strong>
                </div>
              </template>
              <strong v-else class="operator-empty">-</strong>
            </td>

            <td v-if="showCol('namaOperator')" class="operator-cell">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-op-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">
                    {{ op.name || op.label || "-" }}
                    <span
                      v-if="op.shiftTag"
                      class="operator-shift-tag"
                      :class="`shift-tag-${String(op.shiftTagCode || 'NORMAL').toLowerCase()}`"
                    >
                      {{ op.shiftTag }}
                    </span>
                  </strong>
                  <small v-if="op.subText" class="operator-sub">
                    {{ op.subText }}
                  </small>
                </div>
              </template>

              <strong v-else class="operator-name">Not logged in</strong>
            </td>

            <td v-if="showCol('operatorNote')" class="note-cell">
              <template v-if="getOperatorNoteItems(m).length">
                <div
                  v-for="(note, noteIndex) in getOperatorNoteItems(m)"
                  :key="`${m.uuid || index}-note-${note.id || noteIndex}`"
                  class="note-line"
                  :class="{ active: note.isActive }"
                >
                  {{
                    [
                      [note.reasonName, note.body].filter(Boolean).join(" "),
                      note.timeRange,
                    ]
                      .filter(Boolean)
                      .join(" | ") || note.text
                  }}
                </div>
              </template>
              <span v-else>-</span>
            </td>

            <td v-if="showCol('productivity')" class="center productivity-cell">
              {{ Number(m.productivity || 0).toFixed(2) }}%
            </td>

            <td v-if="showCol('status')">
              <span class="badge" :class="statusClass(m.status)">
                {{ m.status || "BAD" }}
              </span>
            </td>

            <td v-if="showCol('runtime')" class="center time-cell">
              {{ formatHour(m.runtime) }}
              <small>{{ formatMinute(m.runtime) }}</small>
            </td>

            <td v-if="showCol('procTime')" class="center time-cell">
              {{ formatHour(m.procTime) }}
              <small>{{ formatMinute(m.procTime) }}</small>
            </td>

            <td v-if="showCol('lossTime')" class="center time-cell">
              {{ formatHour(downTimeSec(m)) }}
              <small>{{ formatMinute(downTimeSec(m)) }}</small>
            </td>

            <td v-if="showCol('output')" class="right">
              <strong>{{ m.output || 0 }}</strong>
              <small>{{ m.complete || 0 }} complete</small>
            </td>

            <td v-if="showCol('uuid')" class="mono">
              {{ m.uuid || "-" }}
              <small>{{ m.tableName || "-" }}</small>
            </td>

            <td v-if="showCol('action')" class="center">
              <button class="btn-edit" type="button" @click="editMachine(m)">
                Edit
              </button>
            </td>

            <td v-if="showCol('ai')" class="center ai-actions">
              <button
                class="btn-ai"
                type="button"
                title="Tanya AI tentang mesin ini"
                @click="askAiMachine(m)"
              >
                Tanya AI
              </button>
              <button
                class="btn-why"
                type="button"
                title="Kenapa jelek"
                @click="whyBadMachine(m)"
              >
                Kenapa
              </button>
            </td>

            <td v-if="showCol('grafik')" class="center">
              <button class="btn-chart" type="button" @click="chartMachine(m)">
                Grafik
              </button>
            </td>
          </template>

          <template v-else>
            <td v-if="showCol('no')" class="right">
              {{ m.__rowNo || props.startIndex + index + 1 }}
            </td>

            <td v-if="showCol('prosesId')" class="col-proses-id">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-user-proses-id-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">{{ getProsesIdFromOperator(op, m) }}</strong>
                </div>
              </template>
              <strong v-else class="operator-name">{{ m.operatorProsesId || "-" }}</strong>
            </td>

            <td v-if="showCol('processName')">
              <strong>{{ getDisplayMachineName(m) }}</strong>
            </td>

            <td v-if="showCol('style')" class="col-style">
              {{ m.operatorStyleName || "-" }}
            </td>

            <td v-if="showCol('branch')">
              {{ getBranchLabel(m) }}
            </td>

            <td v-if="showCol('location')">
              {{ getLocationLineLabel(m) }}
            </td>

            <td v-if="showCol('spvId')" class="col-spv-id" :title="m.spvName || ''">
              {{ m.spvId || "-" }}
            </td>

            <td v-if="showCol('machineId')" class="col-machine-id" :title="m.machineId || ''">
              <div class="machine-id-stack">
                <span
                  v-for="(line, lineIndex) in formatMachineIdLines(m.machineId)"
                  :key="`${m.uuid || index}-user-mid-${lineIndex}`"
                >
                  {{ line }}
                </span>
              </div>
            </td>

            <td v-if="showCol('mikaId')" class="col-mika-id">-</td>

            <td v-if="showCol('nik')" class="col-nik">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-user-nik-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">{{ op.nik || "-" }}</strong>
                </div>
              </template>
              <strong v-else class="operator-empty">-</strong>
            </td>

            <td v-if="showCol('namaOperator')" class="operator-cell">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-user-op-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">
                    {{ op.name || op.label || "-" }}
                    <span
                      v-if="op.shiftTag"
                      class="operator-shift-tag"
                      :class="`shift-tag-${String(op.shiftTagCode || 'NORMAL').toLowerCase()}`"
                    >
                      {{ op.shiftTag }}
                    </span>
                  </strong>
                  <small v-if="op.subText" class="operator-sub">
                    {{ op.subText }}
                  </small>
                </div>
              </template>

              <strong v-else class="operator-name">Not logged in</strong>
            </td>

            <td v-if="showCol('operatorNote')" class="note-cell">
              <template v-if="getOperatorNoteItems(m).length">
                <div
                  v-for="(note, noteIndex) in getOperatorNoteItems(m)"
                  :key="`${m.uuid || index}-note-${note.id || noteIndex}`"
                  class="note-line"
                  :class="{ active: note.isActive }"
                >
                  {{
                    [
                      [note.reasonName, note.body].filter(Boolean).join(" "),
                      note.timeRange,
                    ]
                      .filter(Boolean)
                      .join(" | ") || note.text
                  }}
                </div>
              </template>
              <span v-else>-</span>
            </td>

            <td v-if="showCol('runtime')" class="center time-cell">
              {{ formatHour(m.runtime) }}
              <small>{{ formatMinute(m.runtime) }}</small>
            </td>

            <td v-if="showCol('procTime')" class="center time-cell">
              {{ formatHour(m.procTime) }}
              <small>{{ formatMinute(m.procTime) }}</small>
            </td>

            <td v-if="showCol('lossTime')" class="center time-cell">
              {{ formatHour(downTimeSec(m)) }}
              <small>{{ formatMinute(downTimeSec(m)) }}</small>
            </td>

            <td v-if="showCol('productivity')" class="center productivity-cell">
              {{ Number(m.productivity || 0).toFixed(2) }}%
            </td>

            <td v-if="showCol('status')" class="center">
              <span class="badge" :class="statusClass(m.status)">
                {{ m.status || "BAD" }}
              </span>
            </td>

            <td v-if="showCol('grafik')" class="center">
              <button class="btn-chart" type="button" @click="chartMachine(m)">
                Grafik
              </button>
            </td>
          </template>
        </tr>

        <tr v-if="!props.pagedMachines.length && !props.loading">
          <td :colspan="props.totalCols" class="empty">
            Data tidak ditemukan.
          </td>
        </tr>

        <tr v-if="props.loading">
          <td :colspan="props.totalCols" class="empty">
            Mengambil data...
          </td>
        </tr>
      </tbody>
    </table>

    <div
      v-if="ctxMenu"
      class="row-ctx"
      :style="{ left: `${ctxMenu.x}px`, top: `${ctxMenu.y}px` }"
      @click.stop
    >
      <button type="button" @click="askAiMachine(ctxMenu.machine)">Tanya AI</button>
      <button type="button" @click="whyBadMachine(ctxMenu.machine)">Kenapa jelek</button>
      <button type="button" @click="chartMachine(ctxMenu.machine); closeCtxMenu()">
        Grafik
      </button>
    </div>
  </div>
</template>