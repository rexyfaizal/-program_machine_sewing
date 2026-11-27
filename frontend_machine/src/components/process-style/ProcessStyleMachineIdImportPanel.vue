<script setup>
import { computed, ref } from "vue";
import { downloadMachineIdTemplate } from "../../utils/machineIdImportExcel";

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
  importResult: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["reset-import", "file-change", "submit-import"]);

const templateMessage = ref("");
const templateMessageType = ref("error");

const hasFileStats = computed(
  () => Number(props.importStats?.totalExcelRows || 0) > 0
);

const previewRows = computed(() =>
  (Array.isArray(props.importPreviewRows) ? props.importPreviewRows : []).slice(
    0,
    8
  )
);

function handleDownloadTemplate() {
  templateMessage.value = "";
  templateMessageType.value = "error";

  try {
    const result = downloadMachineIdTemplate(props.templateRows);
    templateMessageType.value = "success";
    templateMessage.value = `Template berhasil (${result.rowCount} baris).`;
  } catch (err) {
    templateMessage.value = err?.message || "Gagal mengunduh template Machine ID.";
  }
}
</script>

<template>
  <section class="import-card machine-id-import-card">
    <div class="import-head">
      <div class="import-head-text">
        <h2>Import Machine ID</h2>
        <p>
          Kolom Excel: <strong>UUID</strong> + <strong>Machine ID</strong>.
          Update cepat ke <strong>machine_setting_manual</strong> tanpa isi Location.
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
            class="btn-primary"
            :disabled="
              !props.isAdmin ||
              props.importing ||
              !Number(props.importStats?.readyRows || 0)
            "
            @click="emit('submit-import')"
          >
            {{ props.importing ? "Mengimpor..." : "Import Machine ID" }}
          </button>
        </div>
      </div>

      <div v-if="templateMessage" class="alert" :class="templateMessageType">
        {{ templateMessage }}
      </div>

      <div v-if="props.importErrorMessage" class="alert error">
        {{ props.importErrorMessage }}
      </div>

      <div v-if="props.importSuccessMessage" class="alert success">
        {{ props.importSuccessMessage }}
      </div>

      <div v-if="hasFileStats" class="import-stats">
        <span>Baris Excel: {{ props.importStats.totalExcelRows || 0 }}</span>
        <span>Siap import: {{ props.importStats.readyRows || 0 }}</span>
        <span>Kosong: {{ props.importStats.skippedEmpty || 0 }}</span>
        <span>Duplikat: {{ props.importStats.skippedDuplicate || 0 }}</span>
        <span>Invalid: {{ props.importStats.skippedInvalid || 0 }}</span>
      </div>

      <div v-if="props.importResult" class="import-stats">
        <span>Berhasil: {{ props.importResult.upserted || 0 }}</span>
        <span>Dilewati: {{ props.importResult.skipped || 0 }}</span>
        <span>Total kirim: {{ props.importResult.total || 0 }}</span>
      </div>

      <div v-if="previewRows.length" class="table-wrap import-preview">
        <table>
          <thead>
            <tr>
              <th>Excel Row</th>
              <th>UUID</th>
              <th>Machine ID</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in previewRows" :key="`${row.uuid}-${row.excelRowNumber}`">
              <td>{{ row.excelRowNumber }}</td>
              <td>{{ row.uuid }}</td>
              <td>
                <strong>{{ row.machineId }}</strong>
              </td>
            </tr>
          </tbody>
        </table>
        <small v-if="props.importPreviewRows.length > previewRows.length">
          Menampilkan {{ previewRows.length }} dari
          {{ props.importPreviewRows.length }} baris siap import.
        </small>
      </div>

      <div
        v-if="Array.isArray(props.importResult?.errors) && props.importResult.errors.length"
        class="alert error"
      >
        <div>
          Contoh error:
          {{
            props.importResult.errors
              .slice(0, 3)
              .map((item) => `baris ${item.excelRow || "-"}: ${item.message}`)
              .join(" | ")
          }}
        </div>
      </div>
    </div>
  </section>
</template>
