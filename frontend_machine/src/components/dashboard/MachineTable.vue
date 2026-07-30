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
});

const emit = defineEmits([
  "edit",
  "update:selectedDate",
  "update:startDate",
  "update:endDate",
  "update:keyword",
  "update:locationValue",
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
      @update:selected-date="emit('update:selectedDate', $event)"
      @update:start-date="emit('update:startDate', $event)"
      @update:end-date="emit('update:endDate', $event)"
      @update:keyword="emit('update:keyword', $event)"
      @update:location-value="emit('update:locationValue', $event)"
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