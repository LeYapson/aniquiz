// Helpers de lecture média.
//
// Les pistes sont stockées sous forme de VIDÉO WebM complète, servie par le
// mirror externe animethemes (https://v.animethemes.moe/{Slug}.webm). En jeu on
// n'a besoin QUE du son : animethemes expose le même extrait en audio-only
// (Opus/Ogg) sur un hôte distinct (https://a.animethemes.moe/{Slug}.ogg),
// typiquement ~3-4 Mo contre la vidéo complète bien plus lourde.
//
// Jouer l'audio-only apporte deux gains :
//   1. Bande passante : on ne télécharge plus la piste vidéo pour rien.
//   2. Connexions : l'audio de jeu passe sur l'hôte « a. » tandis que la vidéo
//      du reveal reste sur « v. » → pools de connexions navigateur séparés
//      (la limite ~6 connexions HTTP/1.1 est par hôte), ce qui réduit encore la
//      pression qui faisait caler la lecture au bout de quelques manches.
//
// ⚠ L'audio-only n'existe pas pour 100 % des pistes (animethemes dérive l'audio
// d'une « source video » choisie par score). L'appelant DOIT donc gérer un repli
// sur l'URL vidéo d'origine si le .ogg renvoie une erreur (404). Voir isAudioOnly.

// audioOnlyUrl transforme une URL vidéo WebM animethemes en son équivalent
// audio-only. Toute URL non reconnue est renvoyée telle quelle (repli sûr).
export function audioOnlyUrl(videoUrl) {
  if (!videoUrl || typeof videoUrl !== "string") return videoUrl;
  if (!videoUrl.includes("://v.animethemes.moe/")) return videoUrl;
  return videoUrl
    .replace("://v.animethemes.moe/", "://a.animethemes.moe/")
    .replace(/\.webm(\?|#|$)/, ".ogg$1");
}

// isAudioOnly indique si une URL pointe vers le flux audio-only dérivé, afin de
// détecter qu'on est déjà sur le repli et éviter une boucle de chargement.
export function isAudioOnly(url) {
  return typeof url === "string" && url.includes("://a.animethemes.moe/");
}
