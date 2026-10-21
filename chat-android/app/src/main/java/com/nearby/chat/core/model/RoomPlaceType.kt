package com.nearby.chat.core.model

import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText

object RoomPlaceType {
    const val COMMUNITY = 1
    const val SCHOOL = 2
    const val BUSINESS = 3
    const val SCENIC = 4
    const val CAMPUS = 5
    const val TRANSIT = 6
    const val OTHER = 7

    data class Option(
        val value: Int,
        @StringRes val label: Int,
        @StringRes val mark: Int,
    )

    val options: List<Option> = listOf(
        Option(COMMUNITY, R.string.place_type_community, R.string.place_type_community_mark),
        Option(SCHOOL, R.string.place_type_school, R.string.place_type_school_mark),
        Option(BUSINESS, R.string.place_type_business, R.string.place_type_business_mark),
        Option(SCENIC, R.string.place_type_scenic, R.string.place_type_scenic_mark),
        Option(CAMPUS, R.string.place_type_campus, R.string.place_type_campus_mark),
        Option(TRANSIT, R.string.place_type_transit, R.string.place_type_transit_mark),
        Option(OTHER, R.string.place_type_other, R.string.place_type_other_mark),
    )

    fun option(placeType: Int): Option = options.firstOrNull { it.value == placeType } ?: options.last()

    @StringRes
    fun label(placeType: Int): Int = option(placeType).label

    fun display(placeType: Int, placeTypeText: String?): UiText {
        val text = placeTypeText?.trim().orEmpty()
        return if (text.isBlank()) UiText.of(label(placeType)) else UiText.of(text)
    }
}
