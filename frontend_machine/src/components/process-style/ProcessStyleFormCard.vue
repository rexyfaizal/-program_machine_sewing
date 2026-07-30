<script setup>
const props = defineProps({
  isAdmin: {
    type: Boolean,
    default: false,
  },
  saving: {
    type: Boolean,
    default: false,
  },
  form: {
    type: Object,
    required: true,
  },
  isEditMode: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:form", "save", "reset"]);

function updateField(field, event) {
  emit("update:form", {
    ...props.form,
    [field]: event.target.value,
  });
}
</script>

<template>
  <section class="form-card">
    <div class="form-title">
      <h2>{{ props.isEditMode ? "Edit Data" : "Tambah Data" }}</h2>

      <button
        v-if="props.isEditMode"
        type="button"
        class="btn-soft"
        @click="emit('reset')"
      >
        Batal Edit
      </button>
    </div>

    <form class="form-grid" @submit.prevent="emit('save')">
      <label>
        <span>Style</span>

        <input
          :value="props.form.styleName"
          type="text"
          placeholder="Contoh: 1101482"
          :disabled="!props.isAdmin || props.saving"
          @input="updateField('styleName', $event)"
        />
      </label>

      <label>
        <span>Proses</span>

        <input
          :value="props.form.processName"
          type="text"
          placeholder="Contoh: Quilting Front Kanan Horizontal"
          :disabled="!props.isAdmin || props.saving"
          @input="updateField('processName', $event)"
        />
      </label>

      <button
        type="submit"
        class="btn-primary btn-save-green"
        :disabled="!props.isAdmin || props.saving"
      >
        {{
          props.saving
            ? "Menyimpan..."
            : props.isEditMode
              ? "Update"
              : "Simpan"
        }}
      </button>
    </form>
  </section>
</template>