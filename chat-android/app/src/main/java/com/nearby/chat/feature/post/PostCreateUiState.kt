package com.nearby.chat.feature.post

import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.ui.UiText

data class PostCreateUiState(
    val editPostId: Long = 0,
    val pageLoading: Boolean = false,
    val loadFailed: Boolean = false,
    val roomId: Long = 0,
    val roomName: String = "",
    val nearbyRooms: List<NearbyRoomItem> = emptyList(),
    val nearbyLoading: Boolean = false,
    val category: String = "",
    val pinned: Boolean = false,
    val title: String = "",
    val content: String = "",
    val placeName: String = "",
    val latitude: Double = 0.0,
    val longitude: Double = 0.0,
    val locationReady: Boolean = false,
    val locating: Boolean = false,
    val locationError: UiText? = null,

    val enableSchedule: Boolean = false,
    val startAt: Long = 0,
    val endAt: Long = 0,
    val enableFee: Boolean = false,
    val feeKind: FeeKind = FeeKind.Charge,
    val feeAmount: String = "",
    val feeDescription: String = "",
    val needSignup: Boolean = false,
    val needCheckIn: Boolean = false,
    val limitParticipants: Boolean = false,
    val maxParticipants: String = "",
    val allowWaitlist: Boolean = false,
    val signupDeadlineAt: Long = 0,

    val enableIdlePrice: Boolean = false,
    val idlePrice: String = "",
    val idleNegotiable: Boolean = true,

    val helpReward: Boolean = false,
    val helpRewardAmount: String = "",
    val helpRewardDesc: String = "",
    val images: List<PostImageItem> = emptyList(),
    val loading: Boolean = false,
    val inForm: Boolean = false,
) {
    val isEdit: Boolean get() = editPostId > 0
    val isRoomPost: Boolean get() = roomId > 0
    val hasUploadingImages: Boolean get() = images.any { it.uploading }
    val imageMediaIds: List<Long> get() = images.map { it.mediaId }.filter { it > 0L }
    val displayNearbyRooms: List<NearbyRoomItem>
        get() {
            if (roomId <= 0) return nearbyRooms
            if (nearbyRooms.any { it.id == roomId }) return nearbyRooms
            if (roomName.isBlank()) return nearbyRooms
            val fallback = NearbyRoomItem(
                id = roomId,
                name = roomName,
                centerLat = latitude,
                centerLng = longitude,
            )
            return listOf(fallback) + nearbyRooms.take(NEARBY_ATTACH_LIMIT - 1)
        }

    companion object {
        const val NEARBY_ATTACH_LIMIT = 5
        const val MAX_IMAGES = 9
    }
}
