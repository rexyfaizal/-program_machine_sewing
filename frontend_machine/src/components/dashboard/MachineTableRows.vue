<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";
import {
  downTimeSec,
  formatHour,
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
});

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
          <th>
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
          <th class="col-proses-id">
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
          <th class="left">
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
          <th>
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
          <th>
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
              <span class="th-sort__label">Location</span>
            </button>
          </th>
          <th class="col-machine-id">
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
          <th class="col-nik">
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
              <span class="th-sort__label">NIK</span>
            </button>
          </th>
          <th>
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
          <th>
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
          <th class="center">
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
          <th>
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
          <th class="center">
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
          <th class="center">
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
          <th class="center">
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
          <th>
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
          <th>
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
          <th class="center">Action</th>
          <th class="center">AI</th>
          <th class="center">Grafik</th>
        </tr>

        <tr v-else>
          <th>
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
          <th class="col-proses-id">
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
          <th class="left">
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
          <th>
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
          <th>
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
              <span class="th-sort__label">Location</span>
            </button>
          </th>
          <th class="col-machine-id">
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
          <th class="col-nik">
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
              <span class="th-sort__label">NIK</span>
            </button>
          </th>
          <th>
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
          <th>
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
          <th class="center">
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
          <th class="center">
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
          <th class="center">
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
          <th class="center">
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
          <th class="center">
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
          <th class="center">Grafik</th>
        </tr>
      </thead>

      <tbody>
        <tr
          v-for="(m, index) in props.pagedMachines"
          :key="`${m.uuid}-${index}`"
          @contextmenu="onRowContextMenu($event, m)"
        >
          <template v-if="props.showActions">
            <td class="right">
              {{ m.__rowNo || props.startIndex + index + 1 }}
            </td>

            <td class="col-proses-id">
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

            <td>
              <strong>{{ getDisplayMachineName(m) }}</strong>

              <small v-if="!isUsingProcessName(m) && m.customName">
                Original: {{ m.originalMachineName || "-" }}
              </small>

              <small
                v-if="m.operatorStyleName && isUsingProcessName(m)"
                class="style-line"
              >
                Style: {{ m.operatorStyleName }}
              </small>
            </td>

            <td>
              <strong>{{ getBranchLabel(m) }}</strong>
            </td>

            <td>
              <strong>{{ getLocationLineLabel(m) }}</strong>
            </td>

            <td class="col-machine-id">
              <strong>{{ m.machineId || "-" }}</strong>
            </td>

            <td class="col-nik">
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

            <td class="operator-cell">
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

            <td class="note-cell">
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
              <small>{{ formatMinute(m.runtime) }}</small>
            </td>

            <td class="center time-cell">
              {{ formatHour(m.procTime) }}
              <small>{{ formatMinute(m.procTime) }}</small>
            </td>

            <td class="center time-cell">
              {{ formatHour(downTimeSec(m)) }}
              <small>{{ formatMinute(downTimeSec(m)) }}</small>
            </td>

            <td class="right">
              <strong>{{ m.output || 0 }}</strong>
              <small>{{ m.complete || 0 }} complete</small>
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

            <td class="center ai-actions">
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

            <td class="center">
              <button class="btn-chart" type="button" @click="chartMachine(m)">
                Grafik
              </button>
            </td>
          </template>

          <template v-else>
            <td class="right">
              {{ m.__rowNo || props.startIndex + index + 1 }}
            </td>

            <td class="col-proses-id">
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

            <td>
              <strong>{{ getDisplayMachineName(m) }}</strong>

              <small
                v-if="m.operatorStyleName && isUsingProcessName(m)"
                class="style-line"
              >
                Style: {{ m.operatorStyleName }}
              </small>
            </td>

            <td>
              {{ getBranchLabel(m) }}
            </td>

            <td>
              {{ getLocationLineLabel(m) }}
            </td>

            <td class="col-machine-id">
              {{ m.machineId || "-" }}
            </td>

            <td class="col-nik">
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

            <td class="operator-cell">
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

            <td class="note-cell">
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

            <td class="center time-cell">
              {{ formatHour(m.runtime) }}
              <small>{{ formatMinute(m.runtime) }}</small>
            </td>

            <td class="center time-cell">
              {{ formatHour(m.procTime) }}
              <small>{{ formatMinute(m.procTime) }}</small>
            </td>

            <td class="center time-cell">
              {{ formatHour(downTimeSec(m)) }}
              <small>{{ formatMinute(downTimeSec(m)) }}</small>
            </td>

            <td class="center productivity-cell">
              {{ Number(m.productivity || 0).toFixed(2) }}%
            </td>

            <td class="center">
              <span class="badge" :class="statusClass(m.status)">
                {{ m.status || "BAD" }}
              </span>
            </td>

            <td class="center">
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