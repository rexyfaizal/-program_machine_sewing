<script setup>
import { toRef } from "vue";
import OperatorLoginForm from "../components/operator/OperatorLoginForm.vue";
import OperatorMenuPanel from "../components/operator/OperatorMenuPanel.vue";
import OperatorConflictPanel from "../components/operator/OperatorConflictPanel.vue";
import { useOperatorMachinePage } from "../composables/useOperatorMachinePage";
import "../assets/operator-machine.css";

const props = defineProps({
  uuid: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["back-dashboard"]);

const page = useOperatorMachinePage(toRef(props, "uuid"));

function backDashboard() {
  emit("back-dashboard");
}
</script>

<template>
  <main class="operator-page">
    <section class="operator-card">
      <div class="card-head">
        <div>
          <p class="eyebrow">QR Operator</p>

          <h1 v-if="page.pageMode.value === 'menu'">
            Menu Keterangan Operator
          </h1>

          <h1 v-else-if="page.pageMode.value === 'conflict'">
            Mesin Sedang Dipakai
          </h1>

          <h1 v-else>
            Login Operator Mesin
          </h1>

          <p v-if="page.pageMode.value === 'menu'">
            Klik keterangan saat terjadi loss time atau kendala proses.
          </p>

        </div>

        <button
        v-if="page.pageMode.value !== 'login'"
        type="button"
        class="btn-back"
        @click="backDashboard"
        >
        Dashboard
        </button>
      </div>

      <div v-if="page.loading.value" class="alert">
        Mengambil data...
      </div>

      <div v-if="page.errorMessage.value" class="alert error">
        {{ page.errorMessage.value }}
      </div>

      <div v-if="page.successMessage.value" class="alert success">
        {{ page.successMessage.value }}
      </div>

      <OperatorLoginForm
        v-if="page.pageMode.value === 'login'"
        v-model:operator-nik="page.operatorNik.value"
        v-model:process-name="page.processName.value"
        v-model:style-name="page.styleName.value"
        :saving="page.saving.value"
        :loading="page.loading.value"
        :force-replace="page.forceReplace.value"
        :operator-name="page.operatorName.value"
        :operator-branch="page.operatorBranch.value"
        :employee-options="page.employeeOptions.value"
        :employee-searching="page.employeeSearching.value"
        :show-employee-options="page.showEmployeeOptions.value"
        :employee-label="page.employeeLabel"
        :style-options="page.styleOptions.value"
        :style-searching="page.styleSearching.value"
        :show-style-options="page.showStyleOptions.value"
        :process-options="page.processOptions.value"
        :process-searching="page.processSearching.value"
        :show-process-options="page.showProcessOptions.value"
        @input-employee="page.handleEmployeeInput"
        @focus-employee="page.showEmployeeOptions.value = page.employeeOptions.value.length > 0"
        @blur-employee="page.hideSuggestionDelay"
        @select-employee="page.selectEmployee"
        @input-style="page.handleStyleInput"
        @focus-style="page.openStyleOptions"
        @blur-style="page.hideStyleSuggestionDelay"
        @select-style="page.selectStyle"
        @input-process="page.handleProcessInput"
        @focus-process="page.openProcessOptions"
        @blur-process="page.hideProcessSuggestionDelay"
        @select-process="page.selectProcess"
        @submit-login="page.submitLogin"
      />

      <OperatorMenuPanel
        v-if="page.pageMode.value === 'menu'"
        v-model:other-note="page.otherNote.value"
        :active-session="page.activeSession.value"
        :machine="page.machine.value"
        :active-notes="page.activeNotes.value"
        :reason-menus="page.reasonMenus"
        :note-saving="page.noteSaving.value"
        :format-date-time="page.formatDateTime"
        :format-time="page.formatTime"
        @submit-note="page.submitNote"
        @login-operator-baru="page.loginOperatorBaru"
      />

      <OperatorConflictPanel
        v-if="page.pageMode.value === 'conflict'"
        :active-session="page.activeSession.value"
        :format-date-time="page.formatDateTime"
        @login-operator-baru="page.loginOperatorBaru"
        @batal="page.batalGantiOperator"
      />
    </section>
  </main>
</template>