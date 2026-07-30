<script setup>
import {
  downTimeSec,
  formatHour,
  getDisplayMachineName,
  getOperatorName,
  getOperatorNote,
  getOperatorSubText,
  isUsingProcessName,
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

const emit = defineEmits(["edit"]);

function editMachine(machine) {
  emit("edit", machine);
}
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
        <tr
          v-for="(m, index) in props.pagedMachines"
          :key="`${m.uuid}-${index}`"
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
  </div>
</template>