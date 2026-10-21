package com.nearby.chat.core.media

import com.nearby.chat.BuildConfig

object MediaUrl {

    fun resolve(url: String?, baseUrl: String = BuildConfig.API_BASE_URL): String {
        if (url.isNullOrBlank()) return ""
        val trimmed = url.trim()
        if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) return trimmed
        if (trimmed.startsWith("/")) return baseUrl.trimEnd('/') + trimmed
        return trimmed
    }
}
