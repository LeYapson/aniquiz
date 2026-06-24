const base = import.meta.env.VITE_API_URL ?? "";

export const API_URL = base;
export const WS_URL = base
  ? base.replace(/^http/, "ws")
  : `ws://${location.host}`;
