package com.nearby.chat.feature.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RoomCreateViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(RoomCreateUiState())
    val uiState: StateFlow<RoomCreateUiState> = _uiState.asStateFlow()

    private val _submitted = Channel<Unit>(Channel.BUFFERED)
    val submitted: Flow<Unit> = _submitted.receiveAsFlow()

    init {
        bootstrapLocation()
    }

    fun onName(v: String) = _uiState.update { it.copy(name = v.take(NAME_MAX)) }

    fun onDesc(v: String) = _uiState.update { it.copy(description = v.take(DESC_MAX)) }

    fun onPlaceType(v: Int) = _uiState.update { it.copy(placeType = v) }

    fun onRadius(v: Float) = _uiState.update {
        it.copy(radius = v.coerceIn(RADIUS_MIN, RADIUS_MAX))
    }

    fun applyLocation(lat: Double, lng: Double, placeName: String) {
        _uiState.update {
            it.copy(
                lat = lat,
                lng = lng,
                placeName = PlaceNames.realOrEmpty(placeName),
                locationReady = true,
                locating = false,
            )
        }
    }

    fun create() {
        val s = _uiState.value
        if (s.name.isBlank()) {
            toaster.fail(R.string.room_create_name_required)
            return
        }
        val placeType = s.placeType
        if (placeType == null) {
            toaster.fail(R.string.room_create_type_required)
            return
        }
        if (!s.locationReady) {
            toaster.fail(R.string.room_create_center_required)
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            runCatching {
                chatRepository.createRoom(
                    name = s.name.trim(),
                    placeType = placeType,
                    centerLat = s.lat,
                    centerLng = s.lng,
                    radius = s.radius.toInt(),
                    description = s.description.trim(),
                )
            }.onSuccess {
                _uiState.update { it.copy(loading = false, name = "", description = "") }
                toaster.success(R.string.room_create_submitted)
                _submitted.send(Unit)
            }.onFailure { e ->
                _uiState.update { it.copy(loading = false) }
                toaster.fail(e, R.string.room_create_failed)
            }
        }
    }

    private fun bootstrapLocation() {
        viewModelScope.launch {
            if (_uiState.value.locationReady) return@launch
            _uiState.update { it.copy(locating = true) }

            val saved = locationRepository.savedLocation()
            if (saved != null) {
                applyLocation(saved.lat, saved.lng, saved.placeName)
                return@launch
            }

            runCatching {
                val loc = locationRepository.currentLocation() ?: return@runCatching
                applyLocation(loc.lat, loc.lng, "")
            }.onFailure {
                _uiState.update { it.copy(locating = false) }
            }
        }
    }

    companion object {
        const val RADIUS_MIN = 100f
        const val RADIUS_MAX = ChatLimits.NEARBY_RADIUS.toFloat()
        private const val NAME_MAX = 32
        private const val DESC_MAX = 100
    }
}
