package com.nearby.chat.core.ui

data class ToastEvent(
    val message: UiText,
    val kind: ToastKind = ToastKind.Info,
)
