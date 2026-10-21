package com.nearby.chat.core.h5

import android.webkit.JavascriptInterface
import androidx.annotation.Keep

@Keep
class NearbyAppBridge(
    private val onClose: () -> Unit,
) {
    @JavascriptInterface
    fun close() {
        onClose()
    }
}
