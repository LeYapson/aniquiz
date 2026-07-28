---
id: 21
date: 28 juillet 2026
datetime: 2026-07-28T12:00:00
tag: Fix
title: "Quiz du jour — plus le même anime plusieurs jours de suite"
---

Un bug faisait que le quiz du jour pouvait tomber sur le **même anime plusieurs jours consécutifs**, car les pistes étaient sélectionnées dans l'ordre de leur ID — et les pistes d'un même anime ont souvent des IDs proches (importées ensemble).

La sélection est maintenant mélangée de façon déterministe : tout le monde reçoit toujours la même piste le même jour, mais les animes sont bien répartis dans la rotation.
