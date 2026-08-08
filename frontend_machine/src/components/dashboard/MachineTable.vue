<script setup>
import MachineTablePagination from "./MachineTablePagination.vue";
import MachineTableRows from "./MachineTableRows.vue";
import MachineTableToolbar from "./MachineTableToolbar.vue";
import { useMachineTable } from "../../composables/useMachineTable";
import "../../assets/machine-table.css";

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
  "edit",
  "chart",
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

const {
  currentPage,
  totalRows,
  totalPages,
  startIndex,
  endIndex,
  pagedMachines,
  totalCols,
  visiblePages,
  nextPage,
  prevPage,
  goToPage,
} = useMachineTable(props);
</script>

<template>
  <section class="machine-table-panel">
    <div class="panel-head modern-head">
      <div>
        <h2>Machine Productivity Details</h2>
      </div>

      <span>{{ totalRows }} data</span>
    </div>

    <MachineTableToolbar
      :loading="props.loading"
      :show-actions="props.showActions"
      :selected-date="props.selectedDate"
      :start-date="props.startDate"
      :end-date="props.endDate"
      :keyword="props.keyword"
      :location-value="props.locationValue"
      :location-options="props.locationOptions"
      :shift-value="props.shiftValue"
      :shift-options="props.shiftOptions"
      :show-shift-filter="props.showShiftFilter"
      :status-value="props.statusValue"
      :status-options="props.statusOptions"
      @update:selected-date="emit('update:selectedDate', $event)"
      @update:start-date="emit('update:startDate', $event)"
      @update:end-date="emit('update:endDate', $event)"
      @update:keyword="emit('update:keyword', $event)"
      @update:location-value="emit('update:locationValue', $event)"
      @update:shift-value="emit('update:shiftValue', $event)"
      @update:status-value="emit('update:statusValue', $event)"
      @date-change="emit('dateChange')"
      @download="emit('download')"
    />

    <MachineTableRows
      :paged-machines="pagedMachines"
      :loading="props.loading"
      :show-actions="props.showActions"
      :start-index="startIndex"
      :total-cols="totalCols"
      @edit="emit('edit', $event)"
      @chart="emit('chart', $event)"
    />

    <MachineTablePagination
      :total-rows="totalRows"
      :start-index="startIndex"
      :end-index="endIndex"
      :current-page="currentPage"
      :total-pages="totalPages"
      :visible-pages="visiblePages"
      @prev="prevPage"
      @next="nextPage"
      @go="goToPage"
    />
  </section>
</template>