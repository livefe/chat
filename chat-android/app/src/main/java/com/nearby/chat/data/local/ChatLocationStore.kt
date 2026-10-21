package com.nearby.chat.data.local

import android.content.Context
import androidx.datastore.preferences.core.doublePreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.model.SavedChatLocation
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.first
import javax.inject.Inject
import javax.inject.Singleton

private val Context.chatLocationDataStore by preferencesDataStore("chat_location_prefs")

@Singleton
class ChatLocationStore @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    private val latKey = doublePreferencesKey("lat")
    private val lngKey = doublePreferencesKey("lng")
    private val placeKey = stringPreferencesKey("place_name")
    private val browsingKey = stringPreferencesKey("browsing")
    private val updatedKey = longPreferencesKey("updated_at")

    suspend fun load(): SavedChatLocation? {
        val prefs = context.chatLocationDataStore.data.first()
        val lat = prefs[latKey] ?: return null
        val lng = prefs[lngKey] ?: return null
        if (lat == 0.0 && lng == 0.0) return null
        return SavedChatLocation(
            lat = lat,
            lng = lng,
            placeName = PlaceNames.realOrEmpty(prefs[placeKey].orEmpty()),
            browsing = prefs[browsingKey] == "1",
            updatedAt = prefs[updatedKey] ?: 0L,
        )
    }

    suspend fun save(lat: Double, lng: Double, placeName: String, browsing: Boolean) {
        context.chatLocationDataStore.edit {
            it[latKey] = lat
            it[lngKey] = lng
            it[placeKey] = PlaceNames.realOrEmpty(placeName)
            it[browsingKey] = if (browsing) "1" else "0"
            it[updatedKey] = System.currentTimeMillis()
        }
    }

    suspend fun clear() {
        context.chatLocationDataStore.edit { it.clear() }
    }
}
