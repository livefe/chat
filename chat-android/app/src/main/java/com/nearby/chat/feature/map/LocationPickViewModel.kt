package com.nearby.chat.feature.map

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.model.GeocodeResult
import com.nearby.chat.core.model.Poi
import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.LocationRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import kotlin.math.abs

@HiltViewModel
class LocationPickViewModel @Inject constructor(
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(LocationPickUiState())
    val uiState: StateFlow<LocationPickUiState> = _uiState.asStateFlow()

    private val _picked = MutableSharedFlow<LocationPickResult>(extraBufferCapacity = 1)
    val picked: SharedFlow<LocationPickResult> = _picked.asSharedFlow()

    private var geocodeJob: Job? = null
    private var searchJob: Job? = null
    private var suppressGeocode = false
    private var geocodedLat = Double.NaN
    private var geocodedLng = Double.NaN

    init {
        viewModelScope.launch { bootstrapCenter() }
    }

    private suspend fun bootstrapCenter() {
        val saved = locationRepository.savedLocation()
        val cachedLoc = locationRepository.peekCachedLocation()
        val target = when {
            saved != null -> MapCenter(saved.lat, saved.lng, 16)
            cachedLoc != null -> MapCenter(cachedLoc.lat, cachedLoc.lng, 16)
            else -> locationRepository.currentLocation(timeoutMs = 5_000)?.let { MapCenter(it.lat, it.lng, 16) }
        }
        if (target == null) {
            suppressGeocode = true
            _uiState.update { it.copy(ready = true) }
            return
        }
        val name = PlaceNames.realOrEmpty(saved?.placeName.orEmpty())
        suppressGeocode = true
        _uiState.update {
            it.copy(
                center = target,
                flyTo = target,
                placeName = name,
                ready = true,
            )
        }
        if (name.isNotBlank()) {
            markGeocoded(target.lat, target.lng)
            return
        }
        if (locationRepository.hasMapKey()) {
            val geo = runCatching { locationRepository.reverseGeocode(target.lat, target.lng) }.getOrNull()
            if (geo != null) {
                markGeocoded(target.lat, target.lng)
                _uiState.update { it.copy(placeName = PlaceNames.realOrEmpty(geo.placeName)) }
                detectRegionFromGeocode(geo)
            }
        }
    }

    fun onCenter(c: MapCenter) {
        if (!_uiState.value.ready) return
        _uiState.update { it.copy(center = c, flyTo = null) }
        if (suppressGeocode) {
            suppressGeocode = false
            return
        }
        if (alreadyGeocoded(c)) return
        reverseGeocode(c)
    }

    private fun markGeocoded(lat: Double, lng: Double) {
        geocodedLat = lat
        geocodedLng = lng
    }

    private fun alreadyGeocoded(c: MapCenter): Boolean {
        if (geocodedLat.isNaN() || geocodedLng.isNaN()) return false
        return abs(c.lat - geocodedLat) < 1e-6 && abs(c.lng - geocodedLng) < 1e-6
    }

    private fun reverseGeocode(c: MapCenter) {
        if (alreadyGeocoded(c)) return
        geocodeJob?.cancel()
        geocodeJob = viewModelScope.launch {
            _uiState.update { it.copy(geocoding = true) }
            delay(350)
            if (alreadyGeocoded(c)) {
                _uiState.update { it.copy(geocoding = false) }
                return@launch
            }
            runCatching { locationRepository.reverseGeocode(c.lat, c.lng) }
                .onSuccess { geo ->
                    markGeocoded(c.lat, c.lng)
                    _uiState.update {
                        it.copy(placeName = PlaceNames.realOrEmpty(geo.placeName), geocoding = false)
                    }
                    if (_uiState.value.selectedRegion == null) {
                        detectRegionFromGeocode(geo)
                    }
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.location_locate_failed)
                    _uiState.update { it.copy(geocoding = false) }
                }
        }
    }

    private suspend fun detectRegionFromGeocode(geo: GeocodeResult) {
        if (_uiState.value.selectedRegion != null) return
        val keyword = geo.city.ifBlank { geo.county }.ifBlank { geo.province }
        if (keyword.isBlank()) return
        val level = when {
            geo.county.isNotBlank() -> 3
            geo.city.isNotBlank() -> 2
            else -> 1
        }
        runCatching {
            locationRepository.searchRegions(keyword = keyword, level = level, limit = 1).firstOrNull()
        }.onSuccess { region ->
            if (region != null) {
                _uiState.update { it.copy(selectedRegion = region) }
            }
        }
    }

    fun selectRegion(region: RegionItem) {
        viewModelScope.launch {
            _uiState.update { it.copy(selectedRegion = region, keyword = "", pois = emptyList()) }
            val center = runCatching {
                locationRepository.regionCenter(region.tiandituGb, region.name)
            }.getOrNull() ?: return@launch
            val target = MapCenter(center.lat, center.lng)
            suppressGeocode = true
            _uiState.update { it.copy(center = target, flyTo = target) }
            runCatching { locationRepository.reverseGeocode(center.lat, center.lng) }
                .onSuccess { geo ->
                    markGeocoded(center.lat, center.lng)
                    _uiState.update { it.copy(placeName = PlaceNames.realOrEmpty(geo.placeName)) }
                }
        }
    }

    fun onKeyword(v: String) {
        _uiState.update { it.copy(keyword = v) }
        searchJob?.cancel()
        if (v.isBlank()) {
            _uiState.update { it.copy(pois = emptyList()) }
            return
        }
        searchJob = viewModelScope.launch {
            delay(300)
            searchNow(v)
        }
    }

    fun search() {
        searchJob?.cancel()
        val v = _uiState.value.keyword.trim()
        if (v.isBlank()) {
            _uiState.update { it.copy(pois = emptyList()) }
            return
        }
        searchJob = viewModelScope.launch { searchNow(v) }
    }

    private suspend fun searchNow(keyword: String) {
        val s = _uiState.value
        val gb = s.selectedRegion?.tiandituGb.orEmpty()
        runCatching {
            if (gb.isNotBlank()) {
                locationRepository.searchPoiInRegion(keyword, gb)
            } else {
                locationRepository.searchPoiNearby(keyword, s.center.lat, s.center.lng)
            }
        }
            .onSuccess { list -> _uiState.update { it.copy(pois = list) } }
            .onFailure { e ->
                toaster.fail(e, R.string.common_search_failed)
                _uiState.update { it.copy(pois = emptyList()) }
            }
    }

    fun clearKeyword() {
        searchJob?.cancel()
        _uiState.update { it.copy(keyword = "", pois = emptyList()) }
    }

    fun selectPoi(poi: Poi) {
        val target = MapCenter(poi.lat, poi.lng)
        suppressGeocode = true
        markGeocoded(poi.lat, poi.lng)
        searchJob?.cancel()
        _uiState.update {
            it.copy(
                center = target,
                flyTo = target,
                placeName = poi.name,
                keyword = "",
                pois = emptyList(),
            )
        }
    }

    fun useCurrentLocation() {
        if (_uiState.value.locating) return
        viewModelScope.launch {
            _uiState.update { it.copy(locating = true) }
            val loc = locationRepository.currentLocation(timeoutMs = 8_000)
            if (loc == null) {
                _uiState.update { it.copy(locating = false) }
                return@launch
            }
            val target = MapCenter(loc.lat, loc.lng)
            suppressGeocode = true
            searchJob?.cancel()
            _uiState.update {
                it.copy(
                    center = target,
                    flyTo = target,
                    selectedRegion = null,
                    keyword = "",
                    pois = emptyList(),
                    locating = false,
                )
            }
            runCatching { locationRepository.reverseGeocode(target.lat, target.lng) }
                .onSuccess { geo ->
                    markGeocoded(target.lat, target.lng)
                    _uiState.update { it.copy(placeName = PlaceNames.realOrEmpty(geo.placeName)) }
                    detectRegionFromGeocode(geo)
                }
                .onFailure { e -> toaster.fail(e, R.string.location_locate_failed) }
        }
    }

    fun confirmPick() {
        val s = _uiState.value
        if (s.locating) return
        val existing = PlaceNames.realOrEmpty(s.placeName)
        if (existing.isNotEmpty()) {
            geocodeJob?.cancel()
            _picked.tryEmit(LocationPickResult(s.center.lat, s.center.lng, existing))
            return
        }
        if (s.geocoding) return
        if (!locationRepository.hasMapKey()) {
            toaster.fail(R.string.location_name_unresolved)
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(geocoding = true) }
            runCatching { locationRepository.reverseGeocode(s.center.lat, s.center.lng) }
                .onSuccess { geo ->
                    val name = PlaceNames.realOrEmpty(geo.placeName)
                    if (name.isEmpty()) {
                        toaster.fail(R.string.location_name_unavailable)
                        _uiState.update { it.copy(geocoding = false, placeName = "") }
                        return@onSuccess
                    }
                    _uiState.update { it.copy(placeName = name, geocoding = false) }
                    _picked.tryEmit(LocationPickResult(s.center.lat, s.center.lng, name))
                }
                .onFailure { e ->
                    toaster.fail(e, R.string.location_locate_failed)
                    _uiState.update { it.copy(geocoding = false) }
                }
        }
    }
}
