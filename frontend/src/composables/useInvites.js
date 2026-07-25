import { ref } from 'vue'
import { API_URL } from '../config'

export function useInvites({ room, authFetch, toast }) {
  const friendsForInvite = ref([])
  const showInvitePicker = ref(false)

  const loadFriendsForInvite = async () => {
    try {
      const res = await authFetch(`${API_URL}/api/friends`)
      if (res.ok) friendsForInvite.value = await res.json()
    } catch { /* silencieux */ }
  }

  const toggleInvitePicker = () => {
    showInvitePicker.value = !showInvitePicker.value
    if (showInvitePicker.value) loadFriendsForInvite()
  }

  const inviteFriend = async (friend) => {
    try {
      const res = await authFetch(`${API_URL}/api/invites`, {
        method: 'POST',
        body: JSON.stringify({ to_user_id: friend.user_id, room_id: room.value }),
      })
      if (res.ok) {
        toast.success(`Invitation envoyée à ${friend.username}`)
        showInvitePicker.value = false
      } else {
        const d = await res.json().catch(() => ({}))
        toast.error(d.error || "Échec de l'invitation")
      }
    } catch {
      toast.error('Erreur réseau')
    }
  }

  return { friendsForInvite, showInvitePicker, toggleInvitePicker, inviteFriend }
}
