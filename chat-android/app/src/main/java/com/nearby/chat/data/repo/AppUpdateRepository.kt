package com.nearby.chat.data.repo

import android.content.Context
import android.content.SharedPreferences
import com.nearby.chat.core.model.AppUpdateInfo
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.AppUpdateApi
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AppUpdateRepository @Inject constructor(
    private val api: AppUpdateApi,
    @ApplicationContext context: Context,
) {
    private val prefs: SharedPreferences =
        context.getSharedPreferences("app_update", Context.MODE_PRIVATE)

    suspend fun check(): AppUpdateInfo = api.check().unwrap().toDomain()

    fun skippedSoftVersion(): String = prefs.getString(KEY_SKIP_SOFT, "").orEmpty()

    fun skipSoftUpdate(latestVersion: String) {
        prefs.edit().putString(KEY_SKIP_SOFT, latestVersion).apply()
    }

    fun lastSoftPromptAt(): Long = prefs.getLong(KEY_LAST_SOFT_AT, 0L)

    fun markSoftPrompted() {
        prefs.edit().putLong(KEY_LAST_SOFT_AT, System.currentTimeMillis()).apply()
    }

    companion object {
        private const val KEY_SKIP_SOFT = "skip_soft_version"
        private const val KEY_LAST_SOFT_AT = "last_soft_prompt_at"
        const val SOFT_PROMPT_INTERVAL_MS = 6L * 60 * 60 * 1000
    }
}
