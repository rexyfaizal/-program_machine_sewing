<script setup>
import { computed } from "vue";

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
});

const emit = defineEmits([
  "update:selectedDate",
  "update:startDate",
  "update:endDate",
  "update:keyword",
  "update:locationValue",
  "update:shiftValue",
  "dateChange",
  "download",
]);

const searchPlaceholder = computed(() => {
  if (props.showActions) {
    return "Search machines, UUID, status, area, location, operator, operator note...";
  }

  return "Search machines, status, area, location, operator logged in, operator note...";
});

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

function updateLocation(event) {
  emit("update:locationValue", event.target.value);
  emit("dateChange");
}

function updateShift(event) {
  emit("update:shiftValue", event.target.value);
  emit("dateChange");
}
</script>

<template>
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
</template>