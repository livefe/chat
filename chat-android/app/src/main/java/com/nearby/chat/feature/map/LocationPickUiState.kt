package com.nearby.chat.feature.map

import com.nearby.chat.core.model.Poi
import com.nearby.chat.core.model.RegionItem

data class LocationPickUiState(
    val center: MapCenter = MapCenter(31.2304, 121.4737),
    val flyTo: MapCenter? = null,
    val placeName: String = "",
    val keyword: String = "",
    val pois: List<Poi> = emptyList(),
    val selectedRegion: RegionItem? = null,
    val ready: Boolean = false,
    val locating: Boolean = false,
    val geocoding: Boolean = false,
)
