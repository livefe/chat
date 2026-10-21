package com.nearby.chat.feature.post.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController

@Composable
internal fun rememberDismissInput(): () -> Unit {
    val focusManager = LocalFocusManager.current
    val keyboard = LocalSoftwareKeyboardController.current
    return remember(focusManager, keyboard) {
        val dismiss: () -> Unit = {
            focusManager.clearFocus(force = true)
            keyboard?.hide()
        }
        dismiss
    }
}
