<script setup>
import { computed, ref } from "vue";
import { downloadOperatorCtStyleTemplate } from "../../utils/operatorCtStyleImportExcel";
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
});

const emit = defineEmits(["reset-import", "file-change", "submit-import"]);

const templateMessage = ref("");
const templateMessageType = ref("error");
const hasFileStats = computed(
  () => Number(props.importStats?.totalExcelRows || 0) > 0
);

function handleDownloadTemplate() {
  templateMessage.value = "";
  templateMessageType.value = "error";

  try {
    const result = downloadOperatorCtStyleTemplate(props.templateRows);
    templateMessageType.value = "success";
    templateMessage.value = `Template berhasil (${result.rowCount} baris).`;
  } catch (err) {
    templateMessage.value =
      err?.message || "Gagal mengunduh template CT GM3.";
  }
}
</script>

<template>
  <section class="import-card import-card-ct">
    <div class="import-head">
      <div class="import-head-text">
        <h2>Upload CT Khusus GM3</h2>
        <p>
          Kolom <strong>STYLE</strong> + <strong>PROSES</strong> +
          <strong>CT TOTAL</strong> · dipakai Kap/Jam GM3.
        </p>
      </div>
    </div>

    <div class="import-body">
      <div class="import-toolbar-row">
        <label class="file-box compact">
          <span>File Excel</span>
          <div class="file-row">
            <input
              :key="props.importInputKey"
              type="file"
              accept=".xls,.xlsx"
              :disabled="!props.isAdmin || props.importing"
              @change="emit('file-change', $event)"
            />
            <small>{{ props.importFileName || "Belum ada file" }}</small>
          </div>
        </label>

        <div class="import-actions">
          <button type="button" class="btn-soft" @click="handleDownloadTemplate">
            Template
          </button>
          <button
            type="button"
            class="btn-soft"
            :disabled="props.importing"
            @click="emit('reset-import')"
          >
            Reset
          </button>
          <button
            type="button"
            class="btn-primary btn-import btn-ct-gm3"
            :disabled="
              !props.isAdmin || props.importing || !props.importPreviewRows.length
            "
            @click="emit('submit-import')"
          >
            {{ props.importing ? "Import..." : "Import CT" }}
          </button>
        </div>
      </div>

      <p v-if="templateMessage" class="alert" :class="templateMessageType">
        {{ templateMessage }}
      </p>

      <div v-if="hasFileStats" class="import-stats">
        <div>
          <strong>{{ props.importStats.totalExcelRows }}</strong>
          <span>Total baris</span>
        </div>
        <div>
          <strong>{{ props.importStats.readyRows }}</strong>
          <span>Siap import</span>
        </div>
        <div>
          <strong>{{ props.importStats.skippedEmpty }}</strong>
          <span>Kosong</span>
        </div>
        <div>
          <strong>{{ props.importStats.skippedDuplicate }}</strong>
          <span>Duplikat</span>
        </div>
      </div>

      <div v-if="props.importErrorMessage" class="alert error">
        {{ props.importErrorMessage }}
      </div>

      <div v-if="props.importSuccessMessage" class="alert success">
        {{ props.importSuccessMessage }}
      </div>

      <div v-if="props.importPreviewRows.length" class="preview-box">
        <div class="preview-title">
          <strong>Preview CT GM3</strong>
          <span>
            {{ props.importPreviewDisplayRows.length }}/{{
              props.importPreviewRows.length
            }}
            baris
          </span>
        </div>

        <div class="preview-table-wrap">
          <table>
            <thead>
              <tr>
                <th>No</th>
                <th>Style</th>
                <th>Proses</th>
                <th>CT Total</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in props.importPreviewDisplayRows"
                :key="`${row.styleName}-${row.processName}-${index}`"
              >
                <td>{{ index + 1 }}</td>
                <td>
                  <strong>{{ row.styleName }}</strong>
                </td>
                <td>{{ row.processName }}</td>
                <td>{{ formatCtNumber(row.ctTotal) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <small v-if="props.importPreviewRows.length > 10">
          Preview 10 data pertama.
        </small>
      </div>

      <div v-if="props.importDuplicateRows.length" class="duplicate-box">
        <div class="preview-title">
          <strong>Duplikat dalam file</strong>
          <span>{{ props.importDuplicateRows.length }} data</span>
        </div>

        <div class="preview-table-wrap">
          <table>
            <thead>
              <tr>
                <th>No</th>
                <th>Row</th>
                <th>Style</th>
                <th>Proses</th>
                <th>Dari</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in props.importDuplicateRows"
                :key="`${row.styleName}-${row.processName}-${row.excelRowNumber}`"
              >
                <td>{{ index + 1 }}</td>
                <td>
                  <strong>{{ row.excelRowNumber }}</strong>
                </td>
                <td>{{ row.styleName }}</td>
                <td>{{ row.processName }}</td>
                <td>{{ row.duplicateOfRowNumber }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
</template>
