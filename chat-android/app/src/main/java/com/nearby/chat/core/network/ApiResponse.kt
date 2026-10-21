package com.nearby.chat.core.network

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class ApiResponse<T>(
    val code: Int,
    val message: String = "",
    val data: T? = null,
)

fun <T> ApiResponse<T>.unwrap(): T {
    if (code != 0) throw ApiException(code, message)
    @Suppress("UNCHECKED_CAST")
    return data as T
}

fun ApiResponse<*>.ensureSuccess() {
    if (code != 0) throw ApiException(code, message)
}

private val apiErrorJson = Json {
    ignoreUnknownKeys = true
}

@Serializable
private data class ApiErrorBody(
    val message: String = "",
)

fun parseApiErrorMessage(raw: String?): String? {
    val text = raw?.trim().orEmpty()
    if (text.isEmpty()) return null
    return runCatching {
        apiErrorJson.decodeFromString(ApiErrorBody.serializer(), text).message.trim()
    }.getOrNull()?.takeIf { it.isNotEmpty() }
}
