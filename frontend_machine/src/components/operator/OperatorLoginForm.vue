<script setup>
defineProps({
  saving: Boolean,
  loading: Boolean,
  forceReplace: Boolean,

  operatorName: String,
  operatorBranch: String,

  employeeOptions: {
    type: Array,
    default: () => [],
  },
  employeeSearching: Boolean,
  showEmployeeOptions: Boolean,
  employeeLabel: {
    type: Function,
    required: true,
  },

  styleOptions: {
    type: Array,
    default: () => [],
  },
  styleSearching: Boolean,
  showStyleOptions: Boolean,

  processOptions: {
    type: Array,
    default: () => [],
  },
  processSearching: Boolean,
  showProcessOptions: Boolean,
});

const operatorNik = defineModel("operatorNik", {
  type: String,
  default: "",
});

const processName = defineModel("processName", {
  type: String,
  default: "",
});

const styleName = defineModel("styleName", {
  type: String,
  default: "",
});

const emit = defineEmits([
  "input-employee",
  "focus-employee",
  "blur-employee",
  "select-employee",

  "input-style",
  "focus-style",
  "blur-style",
  "select-style",

  "input-process",
  "focus-process",
  "blur-process",
  "select-process",

  "submit-login",
]);
</script>

<template>
  <form class="form" @submit.prevent="emit('submit-login')">
    <label class="field">
      <span>NIK Operator</span>

      <div class="suggest-wrap">
        <input
          v-model="operatorNik"
          type="text"
          placeholder="Ketik NIK"
          autocomplete="off"
          @input="emit('input-employee')"
          @focus="emit('focus-employee')"
          @blur="emit('blur-employee')"
        />

        <div
          v-if="showEmployeeOptions && (employeeSearching || employeeOptions.length)"
          class="suggest-box"
        >
          <div v-if="employeeSearching" class="suggest-loading">
            Mencari operator...
          </div>

          <button
            v-for="emp in employeeOptions"
            :key="`emp-${emp.nik}-${emp.name}`"
            type="button"
            class="suggest-item"
            @pointerdown.prevent="emit('select-employee', emp)"
          >
            <strong>{{ emp.nik }}</strong>
            <span>{{ emp.name }}</span>
            <small v-if="emp.branchdetail">{{ emp.branchdetail }}</small>
          </button>
        </div>
      </div>
    </label>

    <label class="field">
      <span>Nama Operator</span>
      <input
        :value="operatorName"
        type="text"
        placeholder="Nama otomatis setelah pilih NIK"
        readonly
      />
    </label>

    <label v-if="operatorBranch" class="field">
      <span>Branch</span>
      <input :value="operatorBranch" type="text" readonly />
    </label>

    <label class="field">
      <span>Style</span>

      <div class="suggest-wrap">
        <input
          v-model="styleName"
          type="text"
          placeholder="Ketik style, contoh: 1101482"
          autocomplete="off"
          @input="emit('input-style')"
          @focus="emit('focus-style')"
          @blur="emit('blur-style')"
        />

        <div
          v-if="showStyleOptions && (styleSearching || styleOptions.length)"
          class="suggest-box"
        >
          <div v-if="styleSearching" class="suggest-loading">
            Mencari style...
          </div>

          <button
            v-for="item in styleOptions"
            :key="`style-${item.styleName}`"
            type="button"
            class="suggest-item style-only"
            @pointerdown.prevent="emit('select-style', item)"
          >
            <strong>{{ item.styleName }}</strong>
          </button>
        </div>
      </div>
    </label>

    <label class="field">
      <span>Proses</span>

      <div class="suggest-wrap">
        <input
          v-model="processName"
          type="text"
          placeholder="Pilih proses sesuai style"
          autocomplete="off"
          :disabled="!styleName"
          @input="emit('input-process')"
          @focus="emit('focus-process')"
          @blur="emit('blur-process')"
        />

        <div
          v-if="showProcessOptions && (processSearching || processOptions.length)"
          class="suggest-box"
        >
          <div v-if="processSearching" class="suggest-loading">
            Mencari proses...
          </div>

          <button
            v-for="item in processOptions"
            :key="`process-${item.id}-${item.processName}`"
            type="button"
            class="suggest-item"
            @pointerdown.prevent="emit('select-process', item)"
          >
            <strong>{{ item.styleName }}</strong>
            <span>{{ item.processName }}</span>
          </button>
        </div>
      </div>
    </label>

    <div v-if="forceReplace" class="replace-note">
      Login ini akan mengganti operator aktif sebelumnya.
    </div>

    <button class="btn-save" type="submit" :disabled="saving || loading">
      {{ saving ? "Login..." : "Login Operator" }}
    </button>
  </form>
</template>