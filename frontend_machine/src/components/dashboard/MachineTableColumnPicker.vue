<script setup>
import { onBeforeUnmount, onMounted } from "vue";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  columns: {
    type: Array,
    default: () => [],
  },
  visibility: {
    type: Object,
    default: () => ({}),
  },
  hiddenCount: {
    type: Number,
    default: 0,
  },
});

const emit = defineEmits(["toggle-open", "close", "toggle-column", "show-all"]);

function onDocClick(event) {
  if (!props.open) return;
  const root = event.target?.closest?.(".column-picker");
  if (!root) emit("close");
}

function onKey(event) {
  if (event.key === "Escape") emit("close");
}

onMounted(() => {
  document.addEventListener("click", onDocClick);
  document.addEventListener("keydown", onKey);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", onDocClick);
  document.removeEventListener("keydown", onKey);
});
</script>

<template>
  <div class="column-picker" @click.stop>
    <button
      type="button"
      class="column-picker__btn"
      :class="{ active: props.open || props.hiddenCount > 0 }"
      @click.stop="emit('toggle-open')"
    >
      Kolom
      <span v-if="props.hiddenCount > 0" class="column-picker__badge">
        -{{ props.hiddenCount }}
      </span>
    </button>

    <div v-if="props.open" class="column-picker__menu" role="menu">
      <div class="column-picker__head">
        <strong>Tampilkan kolom</strong>
        <button type="button" class="column-picker__reset" @click="emit('show-all')">
          Tampilkan semua
        </button>
      </div>

      <label
        v-for="col in props.columns"
        :key="col.key"
        class="column-picker__item"
      >
        <input
          type="checkbox"
          :checked="props.visibility[col.key] !== false"
          @change="emit('toggle-column', col.key)"
        />
        <span>{{ col.label }}</span>
      </label>
    </div>
  </div>
</template>
