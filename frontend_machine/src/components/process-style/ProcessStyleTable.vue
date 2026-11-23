<script setup>
import ProcessStylePagination from "./ProcessStylePagination.vue";
import { formatCtNumber } from "../../utils/operatorCt";

const props = defineProps({
  loading: {
    type: Boolean,
    default: false,
  },
  rows: {
    type: Array,
    default: () => [],
  },
  keyword: {
    type: String,
    default: "",
  },
  source: {
    type: String,
    default: "gcc-live",
  },
  sourceMessage: {
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
  "schedule-search",
  "prev",
  "next",
  "go",
]);

function formatNum(value) {
  const n = Number(value);
  if (!Number.isFinite(n)) return "-";
  return formatCtNumber(n);
}

function formatSmv(value) {
  const n = Number(value);
  if (!Number.isFinite(n)) return "-";
  return String(Number(n.toFixed(6)));
}
</script>

<template>
  <section class="table-card">
    <div class="table-toolbar">
      <div class="table-title-area">
        <div>
          <h2>
            {{ props.keyword ? "Hasil Pencarian Master IE" : "List Master IE (Lokal GCC)" }}
          </h2>
          <p>
            Sumber: tabel lokal <strong>gcc_smv_process</strong>.
            Menampilkan {{ props.totalRows ? props.startIndex + 1 : 0 }}
            - {{ props.endIndex }} dari {{ props.totalRows }} data.
          </p>
          <p v-if="props.sourceMessage" class="source-note">
            {{ props.sourceMessage }}
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
          placeholder="Cari style / proses / proses id / buyer / mesin..."
          @input="
            emit('update:keyword', $event.target.value);
            emit('schedule-search');
          "
          @keyup.enter="emit('search')"
        />

        <button type="button" class="btn-primary small" @click="emit('search')">
          Cari
        </button>
      </div>
    </div>

    <div v-if="props.loading" class="loading-box">
      Mengambil data lokal...
    </div>

    <div v-else class="table-wrap">
      <table class="gcc-table">
        <thead>
          <tr>
            <th>No</th>
            <th>Tanggal</th>
            <th>Buyer</th>
            <th>Style</th>
            <th>Item</th>
            <th>Nama Proses</th>
            <th>Proses ID</th>
            <th>Cycle Time</th>
            <th>SMV Minute</th>
            <th>Output/H</th>
            <th>Mesin</th>
          </tr>
        </thead>

        <tbody>
          <tr v-if="!props.rows.length">
            <td colspan="11" class="empty">
              {{
                props.keyword
                  ? "Tidak ketemu. Coba kata kunci lain atau Sinkron ulang."
                  : "Cache kosong. Klik Sinkron GCC → Tabel Lokal dulu."
              }}
            </td>
          </tr>

          <tr
            v-for="(row, index) in props.rows"
            :key="`${row.headerId}-${row.style}-${row.namaProses}-${index}`"
          >
            <td>{{ index + 1 }}</td>
            <td>{{ row.tanggal || "-" }}</td>
            <td>{{ row.buyer || "-" }}</td>
            <td>
              <strong>{{ row.style || "-" }}</strong>
            </td>
            <td>{{ row.item || "-" }}</td>
            <td>{{ row.namaProses || "-" }}</td>
            <td>{{ row.prosesId || "-" }}</td>
            <td class="ct-cell">{{ formatNum(row.cycleTime) }}</td>
            <td class="ct-cell">{{ formatSmv(row.smvMinute) }}</td>
            <td class="ct-cell">{{ formatNum(row.outputPerH) }}</td>
            <td>{{ row.mesin || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <ProcessStylePagination
      v-if="!props.loading && props.totalRows"
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
