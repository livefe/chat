package com.nearby.chat.core.ui

import androidx.annotation.StringRes
import com.nearby.chat.R
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class Toaster @Inject constructor() {
    private val _events = MutableSharedFlow<ToastEvent>(
        extraBufferCapacity = 8,
        onBufferOverflow = BufferOverflow.DROP_OLDEST,
    )
    val events: SharedFlow<ToastEvent> = _events.asSharedFlow()

    fun show(text: UiText) = emit(text, ToastKind.Info)

    fun show(@StringRes id: Int, vararg args: Any) = emit(UiText.Res(id, args.toList()), ToastKind.Info)

    fun success(text: UiText) = emit(text, ToastKind.Success)

    fun success(@StringRes id: Int, vararg args: Any) = emit(UiText.Res(id, args.toList()), ToastKind.Success)

    fun fail(text: UiText) = emit(text, ToastKind.Fail)

    fun fail(@StringRes id: Int, vararg args: Any) = emit(UiText.Res(id, args.toList()), ToastKind.Fail)

    fun fail(error: Throwable?, @StringRes fallback: Int = R.string.error_operation_failed) {
        if (error is CancellationException) return
        fail(error.toUiText(fallback))
    }

    private fun emit(text: UiText, kind: ToastKind) {
        if (text is UiText.Plain && text.value.isBlank()) return
        _events.tryEmit(ToastEvent(text, kind))
    }
}
