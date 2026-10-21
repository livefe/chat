package com.nearby.chat.data.local

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.util.UUID
import java.util.concurrent.atomic.AtomicReference
import javax.inject.Inject
import javax.inject.Singleton

private val Context.deviceDataStore by preferencesDataStore("device_prefs")

@Singleton
class DeviceId @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    private val key = stringPreferencesKey("device_uuid")
    private val mutex = Mutex()

    private val cached = AtomicReference<String?>(null)

    fun getBlocking(): String = cached.get() ?: runBlocking { get() }

    suspend fun get(): String = cached.get() ?: mutex.withLock { cached.get() ?: load() }

    private suspend fun load(): String {
        var result = ""
        context.deviceDataStore.edit { prefs ->
            result = prefs[key]?.takeIf { it.isNotBlank() } ?: UUID.randomUUID().toString().also { prefs[key] = it }
        }
        cached.set(result)
        return result
    }
}
