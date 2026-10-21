package com.nearby.chat.core.format

import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText
import kotlin.math.roundToInt

fun formatDistance(meters: Double): UiText = if (meters < 1000) {
    UiText.of(R.string.distance_m, meters.roundToInt())
} else {
    UiText.of(R.string.distance_km, "%.1f".format(meters / 1000))
}
