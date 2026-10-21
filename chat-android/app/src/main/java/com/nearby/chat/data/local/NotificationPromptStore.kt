package com.nearby.chat.data.local

import android.content.Context
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.first
import javax.inject.Inject
import javax.inject.Singleton

private val Context.notificationPromptDataStore by preferencesDataStore("notification_prompt_prefs")

@Singleton
class NotificationPromptStore @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    private val promptedKey = booleanPreferencesKey("prompted")

    suspend fun prompted(): Boolean = context.notificationPromptDataStore.data.first()[promptedKey] == true

    suspend fun markPrompted() {
        context.notificationPromptDataStore.edit { it[promptedKey] = true }
    }
}
