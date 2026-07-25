---
id: 13
date: 26 juillet 2026
datetime: 2026-07-26
tag: Fix
title: "Vote skip : le round suivant ne se coupait plus au bout de 5 secondes"
---

Le bouton **⏭ Passer la révélation** (et le vote pour passer l'extrait) pouvaient provoquer un bug : le round suivant se terminait tout seul au bout de quelques secondes, comme si quelqu'un avait déjà voté pour le passer.

C'est corrigé ! Voter pour sauter une manche n'a désormais plus aucun effet sur les manches suivantes. Chaque round repart proprement sur son propre timer. 🎯
