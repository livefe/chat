package com.nearby.chat.core.format

object PlaceNames {
    const val NEARBY_WIRE_VALUE = "附近"

    fun realOrEmpty(name: String): String {
        val n = name.trim()
        return if (n.isEmpty() || n == NEARBY_WIRE_VALUE) "" else n
    }

    fun display(name: String, isCurrentGps: Boolean, nearbyLabel: String): String {
        val real = realOrEmpty(name)
        if (real.isNotEmpty()) return real
        return if (isCurrentGps) nearbyLabel else ""
    }
}
