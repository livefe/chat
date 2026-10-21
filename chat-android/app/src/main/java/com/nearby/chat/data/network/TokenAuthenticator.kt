package com.nearby.chat.data.network

import com.nearby.chat.data.local.TokenStore
import okhttp3.Authenticator
import okhttp3.Request
import okhttp3.Response
import okhttp3.Route
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenAuthenticator @Inject constructor(
    private val tokenStore: TokenStore,
    private val refreshCoordinator: TokenRefreshCoordinator,
) : Authenticator {

    override fun authenticate(route: Route?, response: Response): Request? {
        if (responseCount(response) >= 2) return null
        val path = response.request.url.encodedPath
        if (path.contains("/api/auth/login") || path.contains("/api/auth/refresh")) {
            return null
        }

        val staleToken = response.request.header("Authorization")
            ?.removePrefix("Bearer ")
            ?.trim()
        if (!refreshCoordinator.refreshBlocking(staleToken)) return null
        val newToken = tokenStore.accessTokenOrNull() ?: return null
        return response.request.newBuilder()
            .header("Authorization", "Bearer $newToken")
            .build()
    }

    private fun responseCount(response: Response): Int {
        var result = 1
        var prior = response.priorResponse
        while (prior != null) {
            result++
            prior = prior.priorResponse
        }
        return result
    }
}
