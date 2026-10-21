package com.nearby.chat.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import com.nearby.chat.core.model.UserInfo
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.mapper.toDto
import com.nearby.chat.data.remote.dto.UserInfoDto
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenStore @Inject constructor(
    @ApplicationContext context: Context,
    private val json: Json,
) {
    private val prefs = EncryptedSharedPreferences.create(
        context,
        "session_prefs_enc",
        MasterKey.Builder(context).setKeyScheme(MasterKey.KeyScheme.AES256_GCM).build(),
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
    )

    private val _accessTokenFlow = MutableStateFlow(prefs.getString(KEY_ACCESS, null))
    val accessTokenFlow: Flow<String?> = _accessTokenFlow.asStateFlow()
    val hasToken: Flow<Boolean> = accessTokenFlow.map { !it.isNullOrBlank() }

    private val _userFlow = MutableStateFlow(readUser())
    val userFlow: Flow<UserInfo?> = _userFlow.asStateFlow()

    @Volatile
    private var cachedRefresh: String? = prefs.getString(KEY_REFRESH, null)

    fun accessTokenOrNull(): String? = _accessTokenFlow.value

    fun userOrNull(): UserInfo? = _userFlow.value

    suspend fun accessToken(): String? = accessTokenFlow.first()
    fun refreshToken(): String? = cachedRefresh
    suspend fun user(): UserInfo? = userFlow.first()

    suspend fun saveTokens(access: String, refresh: String) {
        cachedRefresh = refresh
        _accessTokenFlow.value = access
        persist {
            putString(KEY_ACCESS, access)
            putString(KEY_REFRESH, refresh)
        }
    }

    suspend fun saveSession(access: String, refresh: String, user: UserInfo) {
        val raw = json.encodeToString(user.toDto())
        cachedRefresh = refresh
        _accessTokenFlow.value = access
        _userFlow.value = user
        persist {
            putString(KEY_ACCESS, access)
            putString(KEY_REFRESH, refresh)
            putString(KEY_USER, raw)
        }
    }

    suspend fun saveUser(user: UserInfo) {
        val raw = json.encodeToString(user.toDto())
        _userFlow.value = user
        persist { putString(KEY_USER, raw) }
    }

    suspend fun clear() {
        cachedRefresh = null
        _accessTokenFlow.value = null
        _userFlow.value = null
        persist { clear() }
    }

    private suspend fun persist(block: SharedPreferences.Editor.() -> Unit) {
        withContext(Dispatchers.IO) {
            prefs.edit().apply(block).commit()
        }
    }

    private fun readUser(): UserInfo? = prefs.getString(KEY_USER, null)?.let { raw ->
        runCatching { json.decodeFromString<UserInfoDto>(raw).toDomain() }.getOrNull()
    }

    private companion object {
        const val KEY_ACCESS = "access_token"
        const val KEY_REFRESH = "refresh_token"
        const val KEY_USER = "user_json"
    }
}
