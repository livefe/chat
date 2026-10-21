package com.nearby.chat.core.ui

import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.network.ApiException
import com.nearby.chat.core.network.parseApiErrorMessage
import kotlinx.coroutines.CancellationException
import retrofit2.HttpException
import java.io.IOException
import java.net.ConnectException
import java.net.SocketException
import java.net.SocketTimeoutException
import java.net.UnknownHostException

fun Throwable?.toUiText(@StringRes fallback: Int = R.string.error_operation_failed): UiText {
    val fallbackText = UiText.Res(fallback)
    if (this == null) return fallbackText
    when (this) {
        is CancellationException -> return fallbackText

        is ApiException -> {
            val msg = message?.trim().orEmpty()
            return if (msg.isBlank()) UiText.Res(R.string.error_operation_failed) else UiText.Plain(msg)
        }

        is HttpException -> {
            val body = runCatching { response()?.errorBody()?.string() }.getOrNull()
            val business = parseApiErrorMessage(body)
            if (!business.isNullOrBlank()) return UiText.Plain(business)
            return fallbackText
        }

        is SocketTimeoutException -> return UiText.Res(R.string.error_network_timeout)

        is UnknownHostException, is ConnectException -> return UiText.Res(R.string.error_network_unavailable)

        is SocketException, is IOException -> {
            val msg = message?.trim().orEmpty()
            return when {
                msg.contains("Unable to resolve host", ignoreCase = true) -> UiText.Res(R.string.error_network_unavailable)
                msg.contains("Canceled", ignoreCase = true) -> fallbackText
                msg.contains("timeout", ignoreCase = true) -> UiText.Res(R.string.error_network_timeout)
                msg.contains("Failed to connect", ignoreCase = true) -> UiText.Res(R.string.error_server_unreachable)
                msg.contains("Network is unreachable", ignoreCase = true) -> UiText.Res(R.string.error_network_unavailable)
                else -> fallbackText
            }
        }
    }
    val msg = message?.trim().orEmpty()
    if (msg.isBlank()) return fallbackText

    if (
        msg.startsWith("HTTP ") ||
        msg.length > 80 ||
        msg.contains("Exception") ||
        msg.contains("CLEARTEXT") ||
        msg.contains("StandaloneCoroutine") ||
        msg.contains("cancelled", ignoreCase = true)
    ) {
        return fallbackText
    }
    return UiText.Plain(msg)
}
