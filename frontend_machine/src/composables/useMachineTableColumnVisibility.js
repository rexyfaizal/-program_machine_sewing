import { computed, ref, watch } from "vue";
import {
  countVisibleColumns,
  listPickerColumns,
  loadColumnVisibility,
  normalizeColumnVisibility,
  saveColumnVisibility,
} from "../utils/machineTableColumns";

export function useMachineTableColumnVisibility(props) {
  const columnVisibility = ref(loadColumnVisibility());
  const pickerOpen = ref(false);

  const pickerColumns = computed(() =>
    listPickerColumns(Boolean(props.showActions))
  );

  const hiddenCount = computed(() => {
    return pickerColumns.value.filter(
      (col) => columnVisibility.value[col.key] === false
    ).length;
  });

  const totalCols = computed(() =>
    countVisibleColumns(columnVisibility.value, Boolean(props.showActions))
  );

  function isColumnVisible(key) {
    if (key === "no") return true;
    return columnVisibility.value?.[key] !== false;
  }

  function setColumnVisible(key, visible) {
    if (key === "no") return;
    columnVisibility.value = normalizeColumnVisibility({
      ...columnVisibility.value,
      [key]: Boolean(visible),
    });
  }

  function toggleColumn(key) {
    setColumnVisible(key, !isColumnVisible(key));
  }

  function showAllColumns() {
    columnVisibility.value = normalizeColumnVisibility({});
  }

  function togglePicker() {
    pickerOpen.value = !pickerOpen.value;
  }

  function closePicker() {
    pickerOpen.value = false;
  }

  watch(
    columnVisibility,
    (value) => {
      saveColumnVisibility(value);
    },
    { deep: true }
  );

  return {
    columnVisibility,
    pickerOpen,
    pickerColumns,
    hiddenCount,
    totalCols,
    isColumnVisible,
    setColumnVisible,
    toggleColumn,
    showAllColumns,
    togglePicker,
    closePicker,
  };
}
