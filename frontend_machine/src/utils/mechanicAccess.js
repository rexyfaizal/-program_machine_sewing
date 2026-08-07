const STORAGE_KEY = "mechanicDashboardUnlocked";
const MECHANIC_PASSWORD = "mekanik2026";

export function isMechanicUnlocked() {
  return sessionStorage.getItem(STORAGE_KEY) === "1";
}

export function unlockMechanicAccess(password) {
  const value = String(password || "").trim();
  if (value !== MECHANIC_PASSWORD) {
    return false;
  }

  sessionStorage.setItem(STORAGE_KEY, "1");
  return true;
}

export function lockMechanicAccess() {
  sessionStorage.removeItem(STORAGE_KEY);
}
