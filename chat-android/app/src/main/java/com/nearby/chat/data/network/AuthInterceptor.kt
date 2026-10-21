package com.nearby.chat.data.network

import com.nearby.chat.data.local.TokenStore
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthInterceptor @Inject constructor(
    private val tokenStore: TokenStore,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val original = chain.request()
        val path = original.url.encodedPath
        if (path.contains("/api/auth/login") || path.contains("/api/auth/refresh")) {
            return chain.proceed(original)
        }

        val token = tokenStore.accessTokenOrNull()
        if (token.isNullOrBlank()) {
            return chain.proceed(original)
        }
        val authed = original.newBuilder()
            .header("Authorization", "Bearer $token")
            .build()
        return chain.proceed(authed)
    }
}
