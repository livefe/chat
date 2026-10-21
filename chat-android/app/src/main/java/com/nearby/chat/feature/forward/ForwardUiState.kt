package com.nearby.chat.feature.forward

import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.model.NearbyRoomItem

data class ForwardUiState(
    val tab: ForwardTab = ForwardTab.Dest,
    val locationReady: Boolean = false,
    val lat: Double = 0.0,
    val lng: Double = 0.0,
    val rooms: List<NearbyRoomItem> = emptyList(),
    val chats: List<ConversationItem> = emptyList(),
    val roomsLoading: Boolean = false,
    val chatsLoading: Boolean = false,
    val chatsLoadingMore: Boolean = false,
    val chatsHasMore: Boolean = false,
    val sending: Boolean = false,
    val sharing: Boolean = false,
    val query: String = "",
) {
    val filteredRooms: List<NearbyRoomItem>
        get() {
            val q = query.trim()
            if (q.isEmpty()) return rooms
            return rooms.filter { room ->
                room.name.contains(q, ignoreCase = true) || room.placeTypeText.contains(q, ignoreCase = true)
            }
        }

    val filteredChats: List<ConversationItem>
        get() {
            val q = query.trim()
            if (q.isEmpty()) return chats
            return chats.filter { it.peerNickname.contains(q, ignoreCase = true) }
        }
}
