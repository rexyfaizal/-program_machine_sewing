<script setup>
const props = defineProps({
  totalRows: {
    type: Number,
    default: 0,
  },
  startIndex: {
    type: Number,
    default: 0,
  },
  endIndex: {
    type: Number,
    default: 0,
  },
  currentPage: {
    type: Number,
    default: 1,
  },
  totalPages: {
    type: Number,
    default: 1,
  },
  visiblePages: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["prev", "next", "go"]);
</script>

<template>
  <div class="pagination">
    <div class="page-info">
      <span v-if="props.totalRows">
        Menampilkan {{ props.startIndex + 1 }} - {{ props.endIndex }} dari
        {{ props.totalRows }} data
      </span>

      <span v-else>
        Tidak ada data
      </span>
    </div>

    <div class="page-controls">
      <button @click="emit('prev')" :disabled="props.currentPage <= 1">
        Prev
      </button>

      <button
        v-for="page in props.visiblePages"
        :key="page"
        class="page-number"
        :class="{ active: props.currentPage === page }"
        @click="emit('go', page)"
      >
        {{ page }}
      </button>

      <button
        @click="emit('next')"
        :disabled="props.currentPage >= props.totalPages"
      >
        Next
      </button>
    </div>

    <div class="page-total">
      Page {{ props.currentPage }} / {{ props.totalPages }}
    </div>
  </div>
</template>