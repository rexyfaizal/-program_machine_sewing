import { ref } from "vue";

export function useProductivitySocket(onData) {
  const socket = ref(null);

  // Status dibuat lowercase agar cocok dengan GlobalStatusDots.vue
  // active       = hijau
  // connecting  = orange
  // offline     = merah
  const socketStatus = ref("offline");

  let reconnectTimer = null;
  let manualClose = false;
  let lastDate = "";

  function clearReconnectTimer() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  }

  function buildSocketUrl(date) {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";

    return `${protocol}://${window.location.host}/ws/productivity?date=${encodeURIComponent(
      date || ""
    )}`;
  }

  function closeExistingSocket() {
    if (!socket.value) return;

    try {
      socket.value.onopen = null;
      socket.value.onmessage = null;
      socket.value.onerror = null;
      socket.value.onclose = null;
      socket.value.close();
    } catch (err) {
      console.warn("Gagal menutup WebSocket lama:", err);
    }

    socket.value = null;
  }

  function scheduleReconnect(date) {
    clearReconnectTimer();

    socketStatus.value = "connecting";

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;

      if (!manualClose) {
        connect(date || lastDate);
      }
    }, 5000);
  }

  function connect(date) {
    lastDate = date || lastDate;

    clearReconnectTimer();
    closeExistingSocket();

    manualClose = false;
    socketStatus.value = "connecting";

    const url = buildSocketUrl(lastDate);
    const ws = new WebSocket(url);

    socket.value = ws;

    ws.onopen = () => {
      socketStatus.value = "active";
      console.log("WebSocket productivity active:", url);
    };

    ws.onmessage = (event) => {
      try {
        const json = JSON.parse(event.data);

        if (json?.error) {
          socketStatus.value = "error";
          console.error("WebSocket backend error:", json.error);
          return;
        }

        if (typeof onData === "function") {
          onData(json);
        }

        // Selama message masuk, status tetap aktif
        socketStatus.value = "active";
      } catch (err) {
        socketStatus.value = "error";
        console.error("WebSocket parse error:", err);
      }
    };

    ws.onerror = (err) => {
      socketStatus.value = "error";
      console.error("WebSocket error:", err);
    };

    ws.onclose = (event) => {
      if (socket.value === ws) {
        socket.value = null;
      }

      console.warn("WebSocket closed:", {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
      });

      if (manualClose) {
        socketStatus.value = "offline";
        return;
      }

      scheduleReconnect(lastDate);
    };
  }

  function close() {
    manualClose = true;

    clearReconnectTimer();
    closeExistingSocket();

    socketStatus.value = "offline";
  }

  return {
    socketStatus,
    connect,
    close,
  };
}