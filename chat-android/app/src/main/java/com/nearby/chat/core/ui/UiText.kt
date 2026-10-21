package com.nearby.chat.core.ui

import android.content.Context
import androidx.annotation.StringRes
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext

sealed interface UiText {
    data class Plain(val value: String) : UiText

    data class Res(
        @StringRes val id: Int,
        val args: List<Any> = emptyList(),
    ) : UiText

    data class Ellipsized(
        val text: UiText,
        val maxLength: Int,
    ) : UiText

    fun resolve(context: Context): String = when (this) {
        is Plain -> value

        is Res -> {
            val resolvedArgs = args.map { if (it is UiText) it.resolve(context) else it }
            if (resolvedArgs.isEmpty()) {
                context.getString(id)
            } else {
                context.getString(id, *resolvedArgs.toTypedArray())
            }
        }

        is Ellipsized -> {
            val raw = text.resolve(context)
            if (raw.length > maxLength) raw.take(maxLength) + "…" else raw
        }
    }

    companion object {
        fun of(@StringRes id: Int, vararg args: Any): UiText = Res(id, args.toList())

        fun of(value: String): UiText = Plain(value)
    }
}

@Composable
fun UiText.asString(): String = resolve(LocalContext.current)
