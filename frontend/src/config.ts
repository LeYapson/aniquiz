// In dev, set VITE_API_URL in frontend/.env (e.g. http://localhost:8080) — the Vite
// proxy also forwards /api and /ws. In production the frontend is served by the
// backend on the SAME origin, so when VITE_API_URL is unset we use same-origin
// relative requests (empty base) and derive the WebSocket URL from the page location.
const base = import.meta.env.VITE_API_URL ?? "";

export const API_URL = base;
export const WS_URL = base
  ? base.replace(/^http/, "ws")
  : `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}`;
