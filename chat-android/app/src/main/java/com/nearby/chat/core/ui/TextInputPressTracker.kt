package com.nearby.chat.core.ui

import androidx.compose.runtime.compositionLocalOf

class TextInputPressTracker {
    @Volatile
    var pressed: Boolean = false
}

val LocalTextInputPressTracker = compositionLocalOf<TextInputPressTracker?> { null }
