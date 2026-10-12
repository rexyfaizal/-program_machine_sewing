<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";
import {
  downTimeSec,
  formatHour,
  formatMinute,
  getDisplayMachineName,
  getOperatorDisplayRows,
  getOperatorNoteItems,
  isNoLoginRunning,
  isUsingProcessName,
  noLoginRunningLabel,
  statusClass,
} from "../../utils/machineTableFormat";

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
});

const emit = defineEmits(["edit", "chart", "ask-ai", "why-bad"]);

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
          <th>No</th>
          <th class="left">Machine Name</th>
          <th>IP</th>
          <th>Location</th>
          <th>Operator</th>
          <th>Operator Note</th>
          <th class="center">Utilisasi Mesin</th>
          <th>Status</th>
          <th class="center">
            Power On Duration
            <small class="th-sub">mesin nyala</small>
          </th>
          <th class="center">
            Running Time
            <small class="th-sub">mesin jalan</small>
          </th>
          <th class="center">
            Loss Time
            <small class="th-sub">waktu terbuang</small>
          </th>
          <th>Output</th>
          <th>Abnormal</th>
          <th>Avg / Max Proses</th>
          <th>Alarm</th>
          <th>Program Dominan</th>
          <th>UUID / Table</th>
          <th class="center">Action</th>
          <th class="center">AI</th>
          <th class="center">Grafik</th>
        </tr>

        <tr v-else>
          <th>No</th>
          <th class="left">Machine Name</th>
          <th>Location</th>
          <th>Operator</th>
          <th>Operator Note</th>
          <th class="center">
            Power On Duration
            <small class="th-sub">mesin nyala</small>
          </th>
          <th class="center">
            Running Time
            <small class="th-sub">mesin jalan</small>
          </th>
          <th class="center">
            Loss Time
            <small class="th-sub">waktu terbuang</small>
          </th>
          <th class="center">Utilisasi Mesin</th>
          <th class="center">Status</th>
          <th class="center">Grafik</th>
        </tr>
      </thead>

      <tbody>
        <tr
          v-for="(m, index) in props.pagedMachines"
          :key="`${m.uuid}-${index}`"
          :class="{ 'row-no-login': isNoLoginRunning(m) }"
          @contextmenu="onRowContextMenu($event, m)"
        >
          <template v-if="props.showActions">
            <td class="right">
              {{ props.startIndex + index + 1 }}
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

            <td class="mono">
              {{ m.ip || "-" }}
            </td>

            <td>
              <strong>{{ m.location || "-" }}</strong>
            </td>

            <td class="operator-cell">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-op-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">
                    {{ op.label }}
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

              <div v-else class="operator-empty-wrap">
                <strong class="operator-empty">Not logged in</strong>
                <span v-if="isNoLoginRunning(m)" class="badge-no-login">
                  {{ noLoginRunningLabel(m) }}
                </span>
              </div>
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
              {{ props.startIndex + index + 1 }}
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
              {{ m.location || "-" }}
            </td>

            <td class="operator-cell">
              <template v-if="getOperatorDisplayRows(m).length">
                <div
                  v-for="(op, opIndex) in getOperatorDisplayRows(m)"
                  :key="`${m.uuid || index}-user-op-${opIndex}`"
                  class="operator-item"
                >
                  <strong class="operator-name">
                    {{ op.label }}
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

              <div v-else class="operator-empty-wrap">
                <strong class="operator-empty">Not logged in</strong>
                <span v-if="isNoLoginRunning(m)" class="badge-no-login">
                  {{ noLoginRunningLabel(m) }}
                </span>
              </div>
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