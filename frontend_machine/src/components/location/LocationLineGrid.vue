<script setup>
import LocationLineCard from "./LocationLineCard.vue";
import AddLineCard from "./AddLineCard.vue";

defineProps({
  lines: {
    type: Array,
    required: true,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
  draggingLine: {
    type: String,
    default: "",
  },
  dragOverLine: {
    type: String,
    default: "",
  },
  machinesByLine: {
    type: Function,
    required: true,
  },
});

const emit = defineEmits([
  "addLine",
  "renameLine",
  "deleteLine",
  "addMachine",
  "editMachine",
  "removeMachine",
  "dragStart",
  "dragEnter",
  "dragOver",
  "dropLine",
  "dragEnd",
  "openDetailMachine",
]);
</script>

<template>
  <div class="line-grid">
    <LocationLineCard
    v-for="line in lines"
    :key="line"
    :line="line"
    :machines="machinesByLine(line)"
    :is-admin="isAdmin"
    :dragging-line="draggingLine"
    :drag-over-line="dragOverLine"
    @rename-line="emit('renameLine', $event)"
    @delete-line="emit('deleteLine', $event)"
    @add-machine="emit('addMachine', $event)"
    @edit-machine="emit('editMachine', $event)"
    @remove-machine="emit('removeMachine', $event)"
    @drag-start="emit('dragStart', $event.event, $event.line)"
    @drag-enter="emit('dragEnter', $event)"
    @drag-over="emit('dragOver', $event)"
    @drop-line="emit('dropLine', $event)"
    @drag-end="emit('dragEnd')"
    @open-detail-machine="emit('openDetailMachine', $event)"
    />

    <AddLineCard
      v-if="isAdmin"
      @add-line="emit('addLine')"
    />
  </div>
</template>

<style scoped>
.line-grid {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: 10px;
  align-items: start;
}

@media (max-width: 1500px) {
  .line-grid {
    grid-template-columns: repeat(6, 1fr);
  }
}

@media (max-width: 1100px) {
  .line-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 768px) {
  .line-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .line-grid {
    grid-template-columns: 1fr;
  }
}
</style>