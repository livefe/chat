package com.nearby.chat.feature.map

import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.ui.UiText

data class RegionCityPickerUiState(
    val keyword: String = "",
    val loading: Boolean = false,
    val list: List<RegionItem> = emptyList(),
    val stack: List<RegionItem> = emptyList(),
    val error: UiText? = null,
)
