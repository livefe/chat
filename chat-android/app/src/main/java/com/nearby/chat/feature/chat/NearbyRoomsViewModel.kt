package com.nearby.chat.feature.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.model.LatLng
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

private data class NearbyRoomsStatus(
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val locationReady: Boolean = false,
)

@HiltViewModel
class NearbyRoomsViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val status = MutableStateFlow(NearbyRoomsStatus())

    val uiState: StateFlow<NearbyRoomsUiState> = combine(chatRepository.nearbyRooms, status) { rooms, s ->
        NearbyRoomsUiState(
            rooms = rooms.sortedBy { it.distance },
            loading = s.loading,
            refreshing = s.refreshing,
            locationReady = s.locationReady,
        )
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(STOP_TIMEOUT_MS), NearbyRoomsUiState())

    init {
        viewModelScope.launch {
            val loc = currentLocation()
            if (loc == null) {
                status.update { it.copy(locationReady = false) }
                return@launch
            }
            if (chatRepository.nearbyRooms.value.isNotEmpty()) {
                status.update { it.copy(locationReady = true) }
                return@launch
            }
            status.update { it.copy(locationReady = true, loading = true) }
            runCatching { sync(loc) }
            status.update { it.copy(loading = false) }
        }
    }

    fun refresh() {
        if (status.value.refreshing) return
        viewModelScope.launch {
            val loc = currentLocation() ?: return@launch
            status.update { it.copy(refreshing = true, locationReady = true) }
            runCatching { sync(loc) }
                .onFailure { e -> toaster.fail(e, R.string.chat_refresh_failed) }
            status.update { it.copy(refreshing = false) }
        }
    }

    private suspend fun currentLocation(): LatLng? {
        chatRepository.lastLocationOrNull()?.let { return LatLng(it.first, it.second) }
        val saved = runCatching { locationRepository.savedLocation() }.getOrNull() ?: return null
        return LatLng(saved.lat, saved.lng)
    }

    private suspend fun sync(loc: LatLng) {
        chatRepository.syncLocationForNearby(
            loc.lat,
            loc.lng,
            radius = ChatLimits.NEARBY_RADIUS,
            limit = ChatLimits.CHAT_PAGE_LIMIT,
            withMessages = false,
            publishRooms = true,
        )
    }

    private companion object {
        const val STOP_TIMEOUT_MS = 5_000L
    }
}
