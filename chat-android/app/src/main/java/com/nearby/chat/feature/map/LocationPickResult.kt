package com.nearby.chat.feature.map

data class LocationPickResult(
    val lat: Double,
    val lng: Double,
    val placeName: String,
) {
    fun toSharePayload(): String = "$lat$SEPARATOR$lng$SEPARATOR$placeName"

    companion object {
        private const val SEPARATOR = '\u001f'

        fun parseSharePayload(raw: String?): LocationPickResult? {
            if (raw.isNullOrBlank()) return null
            val parts = raw.split(SEPARATOR, limit = 3)
            if (parts.size < 2) return null
            val lat = parts[0].toDoubleOrNull() ?: return null
            val lng = parts[1].toDoubleOrNull() ?: return null
            if (lat == 0.0 && lng == 0.0) return null
            return LocationPickResult(lat, lng, parts.getOrElse(2) { "" })
        }
    }
}
