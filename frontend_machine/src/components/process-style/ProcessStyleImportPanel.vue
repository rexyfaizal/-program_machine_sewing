<script setup>
const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
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
</script>

<template>
  <section class="import-card">
    <div class="import-head">
      <div>
        <h2>Upload Excel Master IE</h2>

        <p>
          Format Excel: kolom <strong>STYLE</strong> dan
          <strong>NAMA PROSES</strong>. Kolom <strong>LINE</strong>
          diabaikan.
        </p>
      </div>

      <button
        type="button"
        class="btn-soft"
        :disabled="props.importing"
        @click="emit('reset-import')"
      >
        Reset Upload
      </button>
    </div>

    <div class="import-body">
      <div class="import-picker">
        <label class="file-box">
          <span>Pilih File Excel</span>

          <input
            :key="props.importInputKey"
            type="file"
            accept=".xls,.xlsx"
            :disabled="!props.isAdmin || props.importing"
            @change="emit('file-change', $event)"
          />

          <small>
            {{ props.importFileName || "Belum ada file dipilih" }}
          </small>
        </label>

        <button
          type="button"
          class="btn-primary btn-import"
          :disabled="
            !props.isAdmin || props.importing || !props.importPreviewRows.length
          "
          @click="emit('submit-import')"
        >
          {{ props.importing ? "Mengimport..." : "Import Excel" }}
        </button>
      </div>

      <div class="import-stats">
        <div>
          <strong>{{ props.importStats.totalExcelRows }}</strong>
          <span>Total baris Excel</span>
        </div>

        <div>
          <strong>{{ props.importStats.readyRows }}</strong>
          <span>Siap import</span>
        </div>

        <div>
          <strong>{{ props.importStats.skippedEmpty }}</strong>
          <span>Kosong dilewati</span>
        </div>

        <div>
          <strong>{{ props.importStats.skippedDuplicate }}</strong>
          <span>Duplikat file</span>
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
          <strong>Preview Import</strong>

          <span>
            Menampilkan {{ props.importPreviewDisplayRows.length }} dari
            {{ props.importPreviewRows.length }} data siap import
          </span>
        </div>

        <div class="preview-table-wrap">
          <table>
            <thead>
              <tr>
                <th>No</th>
                <th>Style</th>
                <th>Nama Proses</th>
              </tr>
            </thead>

            <tbody>
              <tr
                v-for="(row, index) in props.importPreviewDisplayRows"
                :key="`${row.style}-${row.processName}-${index}`"
              >
                <td>{{ index + 1 }}</td>

                <td>
                  <strong>{{ row.style }}</strong>
                </td>

                <td>{{ row.processName }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <small v-if="props.importPreviewRows.length > 10">
          Preview hanya menampilkan 10 data pertama.
        </small>
      </div>

      <div v-if="props.importDuplicateRows.length" class="duplicate-box">
        <div class="preview-title">
          <strong>Data Duplikat Dalam File</strong>

          <span>
            {{ props.importDuplicateRows.length }} data duplikat dilewati
          </span>
        </div>

        <div class="preview-table-wrap">
          <table>
            <thead>
              <tr>
                <th>No</th>
                <th>Row Excel</th>
                <th>Style</th>
                <th>Nama Proses</th>
                <th>Duplikat Dari Row</th>
              </tr>
            </thead>

            <tbody>
              <tr
                v-for="(row, index) in props.importDuplicateRows"
                :key="`${row.style}-${row.processName}-${row.excelRowNumber}`"
              >
                <td>{{ index + 1 }}</td>

                <td>
                  <strong>{{ row.excelRowNumber }}</strong>
                </td>

                <td>{{ row.style }}</td>

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