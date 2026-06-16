const base = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export const API_URL = base;
export const WS_URL = base.replace(/^http/, "ws");
