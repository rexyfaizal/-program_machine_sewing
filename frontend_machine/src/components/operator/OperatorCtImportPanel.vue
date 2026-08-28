<script setup>
import { computed, ref } from "vue";
import { downloadOperatorCtTemplate } from "../../utils/operatorCtImportExcel";
import { formatCtNumber } from "../../utils/operatorCt";

const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
  },
  templateRows: {
    type: Array,
    default: () => [],
  },
  locationFilter: {
    type: String,
    default: "ALL",
  },
  importing: {
    type: Boolean,
    default: false,
  },
  importInputKey: {
    type: Number,
    default: 0,
  },
  importFileName: {
    type: String,
    default: "",
  },
  importPreviewRows: {
    type: Array,
    default: () => [],
  },
  importPreviewDisplayRows: {
    type: Array,
    default: () => [],
  },
  importDuplicateRows: {
    type: Array,
    default: () => [],
  },
  importErrorRows: {
    type: Array,
    default: () => [],
  },
  importStats: {
    type: Object,
    default: () => ({
      totalExcelRows: 0,
      readyRows: 0,
      skippedEmpty: 0,
      skippedDuplicate: 0,
      skippedInvalid: 0,
    }),
  },
  importErrorMessage: {
    type: String,
    default: "",
  },
  importSuccessMessage: {
    type: String,
    default: "",
  },
  embedded: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["reset-import", "file-change", "submit-import"]);

const hasFileStats = computed(() => Number(props.importStats?.totalExcelRows || 0) > 0);
const templateMessage = ref("");
const templateMessageType = ref("error");

function handleDownloadTemplate() {
  templateMessage.value = "";
  templateMessageType.value = "error";

  try {
    const result = downloadOperatorCtTemplate(props.templateRows, {
      locationFilter: props.locationFilter,
    });
    templateMessageType.value = "success";
    templateMessage.value = `Template berhasil (${result.rowCount} baris).`;
  } catch (err) {
    templateMessage.value =
      err?.message || "Gagal mengunduh template CT.";
  }
}
</script>

<template>
  <section v-if="isAdmin" class="import-panel" :class="{ embedded }">
    <div class="import-toolbar">
      <h3>CT Master</h3>

      <label class="file-picker">
        <input
          :key="importInputKey"
          type="file"
          accept=".xlsx,.xls"
          :disabled="importing"
          @change="emit('file-change', $event)"
        />
        <span class="file-label">{{ importFileName || "Pilih file Excel" }}</span>
      </label>

      <div class="import-actions">
        <button type="button" class="btn-secondary" @click="handleDownloadTemplate">
          Template
        </button>
        <button
          type="button"
          class="btn-secondary"
          :disabled="importing"
          @click="emit('reset-import')"
        >
          Reset
        </button>
        <button
          type="button"
          class="btn-primary btn-ct"
          :disabled="importing || !importPreviewRows.length"
          @click="emit('submit-import')"
        >
          {{ importing ? "Import..." : "Import CT" }}
        </button>
      </div>
    </div>

    <p v-if="templateMessage" class="import-alert" :class="templateMessageType">
      {{ templateMessage }}
    </p>

    <p v-if="hasFileStats" class="import-summary">
      {{ importStats.readyRows }} siap
      <template v-if="importStats.skippedInvalid">
        · {{ importStats.skippedInvalid }} invalid
      </template>
      <template v-if="importStats.skippedDuplicate">
        · {{ importStats.skippedDuplicate }} duplikat
      </template>
    </p>

    <p v-if="importErrorMessage" class="import-alert error">{{ importErrorMessage }}</p>
    <p v-if="importSuccessMessage" class="import-alert success">
      {{ importSuccessMessage }}
    </p>

    <div v-if="importPreviewDisplayRows.length" class="preview-box">
      <div class="preview-table-wrap">
        <table>
          <thead>
            <tr>
              <th>Line</th>
              <th>UUID</th>
              <th>CT SUM</th>
              <th>CT STD</th>
              <th>CT</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in importPreviewDisplayRows"
              :key="`${row.uuid}-${row.line}-${row.excelRowNumber}`"
            >
              <td>{{ row.line }}</td>
              <td class="mono">{{ row.uuid }}</td>
              <td>{{ formatCtNumber(row.ctSum) }}</td>
              <td>{{ formatCtNumber(row.ctStd) }}</td>
              <td>{{ formatCtNumber(row.ctValue) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-if="importPreviewRows.length > 10" class="preview-more">
        +{{ importPreviewRows.length - 10 }} baris lainnya
      </p>
    </div>

    <div v-if="importErrorRows.length" class="preview-box error-box">
      <div class="preview-table-wrap">
        <table>
          <thead>
            <tr>
              <th>Baris</th>
              <th>Line</th>
              <th>UUID</th>
              <th>Pesan</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in importErrorRows.slice(0, 5)" :key="row.excelRowNumber">
              <td>{{ row.excelRowNumber }}</td>
              <td>{{ row.line || "-" }}</td>
              <td class="mono">{{ row.uuid || "-" }}</td>
              <td>{{ row.message }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<style scoped>
.import-panel {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px 14px;
}

.import-panel.embedded {
  border: 0;
  border-radius: 0;
  padding: 0;
  background: transparent;
}

.import-toolbar {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
}

.import-toolbar h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 800;
  color: #0f172a;
}

.file-picker {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border: 1px solid #dbeafe;
  border-radius: 8px;
  background: #f8fafc;
  cursor: pointer;
}

.file-picker input[type="file"] {
  width: 92px;
  font-size: 12px;
}

.file-label {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.import-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.btn-secondary,
.btn-primary {
  border: 0;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}

.btn-secondary {
  background: #f1f5f9;
  color: #334155;
}

.btn-primary.btn-ct {
  background: #2563eb;
  color: #fff;
}

.btn-primary:disabled,
.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.import-summary {
  margin: 8px 0 0;
  font-size: 12px;
  color: #64748b;
}

.import-alert {
  margin: 8px 0 0;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}

.import-alert.error {
  background: #fef2f2;
  color: #b91c1c;
}

.import-alert.success {
  background: #ecfdf5;
  color: #047857;
}

.preview-box {
  margin-top: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.error-box {
  border-color: #fecaca;
}

.preview-table-wrap {
  overflow-x: auto;
}

.preview-table-wrap table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.preview-table-wrap th,
.preview-table-wrap td {
  padding: 6px 8px;
  border-bottom: 1px solid #f1f5f9;
  text-align: left;
}

.preview-table-wrap th {
  background: #f8fafc;
  color: #64748b;
  font-size: 11px;
}

.preview-more {
  margin: 0;
  padding: 6px 8px;
  font-size: 11px;
  color: #64748b;
  background: #f8fafc;
}

.mono {
  font-family: Consolas, monospace;
  font-size: 11px;
}
</style>
