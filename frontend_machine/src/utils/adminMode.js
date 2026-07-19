export function getInitialAdminMode() {
  const params = new URLSearchParams(window.location.search);

  if (params.get("admin") === "1") {
    localStorage.setItem("machineDashboardAdmin", "1");
    return true;
  }

  if (params.get("admin") === "0") {
    localStorage.removeItem("machineDashboardAdmin");
    return false;
  }

  return localStorage.getItem("machineDashboardAdmin") === "1";
}