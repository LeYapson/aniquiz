// Catalogue des cadres d'avatar cosmétiques, débloqués par niveau.
// DOIT rester synchronisé avec frameUnlockLevel côté backend
// (internal/handlers/cosmetics.go).
export const FRAMES = [
  { id: '',         name: 'Aucun',       level: 1 },
  { id: 'bronze',   name: 'Bronze',      level: 2 },
  { id: 'silver',   name: 'Argent',      level: 5 },
  { id: 'gold',     name: 'Or',          level: 10 },
  { id: 'emerald',  name: 'Émeraude',    level: 15 },
  { id: 'sapphire', name: 'Saphir',      level: 20 },
  { id: 'ruby',     name: 'Rubis',       level: 30 },
  { id: 'rainbow',  name: 'Arc-en-ciel', level: 50 },
];

// Classe CSS du cadre (définie globalement dans style.css). '' = pas de cadre.
export function frameClass(id) {
  return id ? `frame-${id}` : '';
}
