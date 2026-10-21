package com.nearby.chat.feature.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.mergeMessages
import com.nearby.chat.core.chat.trimMessages
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.lifecycle.AppVisibility
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import com.nearby.chat.feature.chat.thread.ChatRoute
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadFactory
import com.nearby.chat.feature.chat.thread.ChatThreadState
import com.nearby.chat.feature.chat.thread.messageScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import kotlin.math.abs

private class LocateFailedException : IllegalStateException()

@HiltViewModel
class NearbyChatViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val locationRepository: LocationRepository,
    private val appVisibility: AppVisibility,
    private val toaster: Toaster,
    threadFactory: ChatThreadFactory,
) : ViewModel() {
    private val _uiState = MutableStateFlow(NearbyChatUiState())
    val uiState: StateFlow<NearbyChatUiState> = _uiState.asStateFlow()

    val rooms: StateFlow<List<NearbyRoomItem>> = chatRepository.nearbyRooms

    val thread: ChatThread = threadFactory.create(
        scope = viewModelScope,
        messageScope = MessageScopes.NEARBY,
        accepts = { event, item ->
            event.messageScope == MessageScopes.NEARBY || item.scope == MessageScopes.NEARBY
        },
        route = {
            val s = _uiState.value
            if (s.locationReady) ChatRoute(lat = s.lat, lng = s.lng) else null
        },
    )
    val threadState: StateFlow<ChatThreadState> = thread.state

    private var presencePolling = false
    private var presenceJob: Job? = null

    init {
        viewModelScope.launch {
            chatRepository.reconnectTick.collect {
                val s = _uiState.value
                if (s.locationReady) {
                    runCatching {
                        bootstrapAt(lat = s.lat, lng = s.lng, place = s.placeName, browsing = s.isBrowsing, persist = false)
                    }
                }
            }
        }
    }

    fun setNearbyPresencePolling(active: Boolean) {
        if (presencePolling == active) return
        presencePolling = active
        presenceJob?.cancel()
        presenceJob = null
        if (!active) return
        presenceJob = viewModelScope.launch {
            while (presencePolling) {
                delay(NEARBY_PRESENCE_INTERVAL_MS)
                if (!presencePolling || !appVisibility.isForeground.value) continue
                val s = _uiState.value
                if (!s.locationReady || s.locating || s.loading) continue
                runCatching { refreshNearbyPresenceOnly(s.lat, s.lng) }
            }
        }
    }

    fun loadPreferCached(canUseGps: Boolean) {
        viewModelScope.launch {
            val s = _uiState.value
            if (s.locating || s.loading || s.refreshing) return@launch
            if (s.locationReady) {
                runCatching {
                    if (threadState.value.messages.isNotEmpty()) {
                        refreshNearbyPresenceOnly(s.lat, s.lng)
                    } else {
                        bootstrapAt(lat = s.lat, lng = s.lng, place = s.placeName, browsing = s.isBrowsing, persist = false)
                    }
                }.onFailure { e -> toaster.fail(e, R.string.common_load_failed) }
                return@launch
            }
            _uiState.update { it.copy(loading = true) }

            val saved = locationRepository.savedLocation()
            if (saved != null) {
                runCatching {
                    val needPlace = !saved.browsing && PlaceNames.realOrEmpty(saved.placeName).isEmpty()
                    if (needPlace) {
                        _uiState.update { it.copy(locating = true) }
                    }
                    bootstrapAt(
                        lat = saved.lat,
                        lng = saved.lng,
                        place = saved.placeName,
                        browsing = saved.browsing,
                        persist = false,
                        clearBusyFlags = !needPlace,
                    )
                    if (needPlace) {
                        fillGpsPlaceNameIfNeeded(saved.lat, saved.lng)
                        _uiState.update { it.copy(locating = false, loading = false, refreshing = false) }
                    }
                }.onFailure { e ->
                    _uiState.update { it.copy(loading = false, locating = false) }
                    toaster.fail(e, R.string.common_load_failed)
                }
                return@launch
            }

            if (!canUseGps) {
                _uiState.update { it.copy(loading = false, locating = false) }
                return@launch
            }

            locateAndBootstrap()
        }
    }

    fun onNearbyTabVisible() {
        val s = _uiState.value
        if (!s.locationReady || s.locating || s.loading || s.refreshing) return
        if (threadState.value.messages.isEmpty()) {
            loadPreferCached(canUseGps = false)
            return
        }
        viewModelScope.launch {
            runCatching { refreshNearbyPresenceOnly(s.lat, s.lng) }
        }
    }

    fun refreshGpsAndBootstrap() {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    locating = true,
                    loading = true,
                    locationFailed = false,
                    isBrowsing = false,
                    placeName = "",
                )
            }
            thread.clear()
            locateAndBootstrap()
        }
    }

    fun applyPickedLocation(lat: Double, lng: Double, placeName: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(locating = false, loading = true, isBrowsing = true) }
            thread.clear()
            val place = PlaceNames.realOrEmpty(placeName)
            if (place.isEmpty()) {
                _uiState.update { it.copy(loading = false) }
                toaster.fail(R.string.location_name_unavailable)
                return@launch
            }
            runCatching {
                locationRepository.saveLocation(lat, lng, place, browsing = true)
                chatRepository.rememberLocation(lat, lng)
                _uiState.update {
                    it.copy(
                        lat = lat,
                        lng = lng,
                        placeName = place,
                        isBrowsing = true,
                        locationReady = true,
                        locationFailed = false,
                    )
                }
                bootstrapAt(lat, lng, place, browsing = true, persist = false, replaceMessages = true)
            }.onFailure { e ->
                _uiState.update { it.copy(loading = false) }
                toaster.fail(e, R.string.common_load_failed)
            }
        }
    }

    fun loadMoreMessages() {
        val s = _uiState.value
        if (!s.locationReady) return
        thread.loadOlder { beforeId ->
            chatRepository.nearbyMessages(s.lat, s.lng, beforeId, ChatLimits.NEARBY_RADIUS, ChatLimits.CHAT_PAGE_LIMIT)
        }
    }

    private suspend fun locateAndBootstrap() {
        runCatching {
            _uiState.update { it.copy(locating = true, loading = true, locationFailed = false) }
            val loc = locationRepository.refreshLocation() ?: throw LocateFailedException()
            bootstrapAt(
                loc.lat,
                loc.lng,
                place = "",
                browsing = false,
                persist = true,
                replaceMessages = true,
                clearBusyFlags = false,
            )
            fillGpsPlaceNameIfNeeded(loc.lat, loc.lng)
            _uiState.update { it.copy(locating = false, loading = false, refreshing = false) }
        }.onFailure { e ->
            _uiState.update { it.copy(locating = false, loading = false, locationFailed = true) }
            if (e is LocateFailedException) {
                toaster.fail(R.string.location_locate_failed_hint)
            } else {
                toaster.fail(e, R.string.location_locate_failed)
            }
        }
    }

    private suspend fun bootstrapAt(
        lat: Double,
        lng: Double,
        place: String,
        browsing: Boolean,
        persist: Boolean,
        replaceMessages: Boolean = false,
        clearBusyFlags: Boolean = true,
    ) {
        val name = PlaceNames.realOrEmpty(place)
        val prev = _uiState.value
        val locationChanged = !prev.locationReady || !sameSpot(prev.lat, prev.lng, lat, lng)
        val shouldReplace = replaceMessages || locationChanged

        if (persist) {
            locationRepository.saveLocation(lat, lng, name, browsing)
        }
        chatRepository.rememberLocation(lat, lng)
        _uiState.update {
            it.copy(
                lat = lat,
                lng = lng,
                placeName = name,
                isBrowsing = browsing,
                locationReady = true,
                locationFailed = false,
            )
        }
        val boot = chatRepository.syncLocationForNearby(
            lat,
            lng,
            radius = ChatLimits.NEARBY_RADIUS,
            limit = ChatLimits.CHAT_PAGE_LIMIT,
            withMessages = true,
            publishRooms = true,
        )
        thread.updateMessages(hasMore = boot.messages.size >= ChatLimits.CHAT_PAGE_LIMIT) { current ->
            if (shouldReplace) boot.messages.sortedBy { m -> m.id } else trimMessages(mergeMessages(current, boot.messages))
        }
        _uiState.update {
            val settled = it.copy(onlineCount = boot.onlineCount)
            if (clearBusyFlags) settled.copy(locating = false, loading = false, refreshing = false) else settled
        }
    }

    private suspend fun fillGpsPlaceNameIfNeeded(lat: Double, lng: Double) {
        val s = _uiState.value
        if (s.isBrowsing || !sameSpot(s.lat, s.lng, lat, lng)) return
        if (PlaceNames.realOrEmpty(s.placeName).isNotEmpty()) return
        val resolved = resolvePlaceName(lat, lng)
        if (resolved.isEmpty()) return
        val cur = _uiState.value
        if (cur.isBrowsing || !sameSpot(cur.lat, cur.lng, lat, lng)) return
        locationRepository.saveLocation(lat, lng, resolved, browsing = false)
        _uiState.update { it.copy(placeName = resolved) }
    }

    private suspend fun resolvePlaceName(lat: Double, lng: Double): String {
        if (!locationRepository.hasMapKey()) return ""
        return runCatching { locationRepository.reverseGeocode(lat, lng).placeName }
            .getOrNull()
            ?.let { PlaceNames.realOrEmpty(it) }
            .orEmpty()
    }

    private suspend fun refreshNearbyPresenceOnly(lat: Double, lng: Double) {
        val boot = chatRepository.syncLocationForNearby(
            lat,
            lng,
            radius = ChatLimits.NEARBY_RADIUS,
            limit = ChatLimits.CHAT_PAGE_LIMIT,
            withMessages = false,
            publishRooms = true,
        )
        _uiState.update { it.copy(onlineCount = boot.onlineCount) }
    }

    private fun sameSpot(aLat: Double, aLng: Double, bLat: Double, bLng: Double): Boolean = abs(aLat - bLat) <= COORD_EPSILON && abs(aLng - bLng) <= COORD_EPSILON

    private companion object {
        const val NEARBY_PRESENCE_INTERVAL_MS = 30_000L
        const val COORD_EPSILON = 1e-6
    }
}
