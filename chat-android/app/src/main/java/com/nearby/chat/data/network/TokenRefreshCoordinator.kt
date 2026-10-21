package com.nearby.chat.data.network

import com.nearby.chat.BuildConfig
import com.nearby.chat.core.network.ApiResponse
import com.nearby.chat.core.network.RefreshHttpClient
import com.nearby.chat.data.local.TokenStore
import com.nearby.chat.data.remote.dto.RefreshTokenRequest
import com.nearby.chat.data.remote.dto.RefreshTokenResultDto
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenRefreshCoordinator @Inject constructor(
    @RefreshHttpClient private val refreshClient: OkHttpClient,
    private val tokenStore: TokenStore,
    private val json: Json,
) {
    private val mutex = Mutex()

    fun refreshBlocking(staleAccessToken: String?): Boolean = runBlocking { refresh(staleAccessToken) }

    suspend fun refresh(staleAccessToken: String?): Boolean = mutex.withLock {
        val latest = tokenStore.accessTokenOrNull()
        if (!staleAccessToken.isNullOrBlank() && !latest.isNullOrBlank() && latest != staleAccessToken) {
            return@withLock true
        }
        val refreshToken = tokenStore.refreshToken() ?: return@withLock false
        when (val outcome = callRefresh(refreshToken)) {
            is RefreshOutcome.Ok -> {
                tokenStore.saveTokens(outcome.result.accessToken, outcome.result.refreshToken)
                true
            }

            RefreshOutcome.Invalid -> {
                tokenStore.clear()
                false
            }

            RefreshOutcome.Transient -> false
        }
    }

    private sealed class RefreshOutcome {
        data class Ok(val result: RefreshTokenResultDto) : RefreshOutcome()
        data object Invalid : RefreshOutcome()
        data object Transient : RefreshOutcome()
    }

    private fun callRefresh(refreshToken: String): RefreshOutcome {
        return try {
            val bodyJson = json.encodeToString(
                RefreshTokenRequest.serializer(),
                RefreshTokenRequest(refreshToken),
            )
            val request = Request.Builder()
                .url(BuildConfig.API_BASE_URL.trimEnd('/') + "/api/auth/refresh")
                .post(bodyJson.toRequestBody("application/json".toMediaType()))
                .build()
            refreshClient.newCall(request).execute().use { resp ->
                when (resp.code) {
                    401, 403, 404 -> return RefreshOutcome.Invalid
                }
                if (!resp.isSuccessful) return RefreshOutcome.Transient
                val raw = resp.body?.string() ?: return RefreshOutcome.Transient
                val parsed = json.decodeFromString(
                    ApiResponse.serializer(RefreshTokenResultDto.serializer()),
                    raw,
                )
                if (parsed.code != 0 || parsed.data == null) return RefreshOutcome.Invalid
                RefreshOutcome.Ok(parsed.data)
            }
        } catch (_: IOException) {
            RefreshOutcome.Transient
        } catch (_: Exception) {
            RefreshOutcome.Transient
        }
    }
}
