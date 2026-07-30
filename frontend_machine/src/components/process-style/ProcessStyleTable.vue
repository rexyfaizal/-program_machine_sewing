<script setup>
import ProcessStylePagination from "./ProcessStylePagination.vue";

const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  rows: {
    type: Array,
    default: () => [],
  },
  pagedRows: {
    type: Array,
    default: () => [],
  },
  keyword: {
    type: String,
    default: "",
  },
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

const emit = defineEmits([
  "update:keyword",
  "search",
  "edit",
  "remove",
  "prev",
  "next",
  "go",
]);

function formatDateTime(value) {
  if (!value) return "-";

  const d = new Date(value);

  if (Number.isNaN(d.getTime())) {
    return String(value).replace("T", " ").slice(0, 19);
  }

  return d.toLocaleString("id-ID", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
</script>

<template>
  <section class="table-card">
    <div class="table-toolbar">
      <div class="table-title-area">
        <div>
          <h2>List Master IE</h2>

          <p>
            Menampilkan {{ props.totalRows ? props.startIndex + 1 : 0 }} -
            {{ props.endIndex }} dari {{ props.totalRows }} data.
          </p>
        </div>

        <span class="table-count-badge">
          {{ props.totalRows }} data
        </span>
      </div>

      <div class="search-box">
        <input
          :value="props.keyword"
          type="text"
          placeholder="Cari style / proses..."
          @input="emit('update:keyword', $event.target.value)"
          @keyup.enter="emit('search')"
        />

        <button type="button" class="btn-primary small" @click="emit('search')">
          Refresh
        </button>
      </div>
    </div>

    <div v-if="props.loading" class="loading-box">
      Mengambil data...
    </div>

    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>No</th>
            <th>Style</th>
            <th>Proses</th>
            <th>Created At</th>
            <th v-if="props.isAdmin">Action</th>
          </tr>
        </thead>

        <tbody>
          <tr v-if="!props.rows.length">
            <td :colspan="props.isAdmin ? 5 : 4" class="empty">
              Belum ada data.
            </td>
          </tr>

          <tr v-for="(row, index) in props.pagedRows" :key="row.id">
            <td>{{ props.startIndex + index + 1 }}</td>

            <td>
              <strong>{{ row.styleName }}</strong>
            </td>

            <td>{{ row.processName }}</td>

            <td>{{ formatDateTime(row.createdAt) }}</td>

            <td v-if="props.isAdmin">
              <div class="actions">
                <button type="button" class="btn-edit" @click="emit('edit', row)">
                  Edit
                </button>

                <button
                  type="button"
                  class="btn-delete"
                  @click="emit('remove', row)"
                >
                  Hapus
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <ProcessStylePagination
      v-if="!props.loading && props.rows.length"
      :total-rows="props.totalRows"
      :start-index="props.startIndex"
      :end-index="props.endIndex"
      :current-page="props.currentPage"
      :total-pages="props.totalPages"
      :visible-pages="props.visiblePages"
      @prev="emit('prev')"
      @next="emit('next')"
      @go="emit('go', $event)"
    />
  </section>
</template>