package com.nearby.chat.feature.h5

import android.annotation.SuppressLint
import android.content.Context
import android.graphics.Bitmap
import android.net.Uri
import android.net.http.SslError
import android.webkit.SslErrorHandler
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.nearby.chat.BuildConfig
import com.nearby.chat.R
import com.nearby.chat.core.h5.NearbyAppBridge
import com.nearby.chat.core.ui.UiText

@Stable
class H5WebController(
    context: Context,
    private val fixedTitle: String,
    defaultTitle: String,
) {
    var pageTitle by mutableStateOf(fixedTitle.ifBlank { defaultTitle })
        private set
    var progress by mutableFloatStateOf(0f)
        private set
    var loadError by mutableStateOf<UiText?>(null)
        private set

    @SuppressLint("SetJavaScriptEnabled")
    val webView: WebView = WebView(context).apply {
        settings.javaScriptEnabled = true
        settings.domStorageEnabled = true
        settings.cacheMode = WebSettings.LOAD_DEFAULT
        settings.mixedContentMode = if (BuildConfig.DEBUG) {
            WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE
        } else {
            WebSettings.MIXED_CONTENT_NEVER_ALLOW
        }
    }

    private var currentUrl = ""

    fun attach(onClose: () -> Unit) {
        webView.addJavascriptInterface(NearbyAppBridge(onClose = onClose), BRIDGE_NAME)
        webView.webChromeClient = chromeClient
        webView.webViewClient = viewClient
    }

    fun release() {
        webView.removeJavascriptInterface(BRIDGE_NAME)
        webView.stopLoading()
        webView.destroy()
    }

    fun load(url: String) {
        currentUrl = url
        if (url.isBlank()) {
            loadError = UiText.of(R.string.h5_url_empty)
            return
        }
        loadError = null
        webView.loadUrl(url)
    }

    fun retry() {
        loadError = null
        if (currentUrl.isNotBlank()) webView.loadUrl(currentUrl)
    }

    fun goBack(): Boolean {
        if (!webView.canGoBack()) return false
        webView.goBack()
        return true
    }

    private val chromeClient = object : WebChromeClient() {
        override fun onProgressChanged(view: WebView?, newProgress: Int) {
            progress = newProgress / PERCENT
        }

        override fun onReceivedTitle(view: WebView?, t: String?) {
            if (fixedTitle.isBlank() && !t.isNullOrBlank() && !t.startsWith("http")) {
                pageTitle = t
            }
        }
    }

    private val viewClient = object : WebViewClient() {
        override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
            val target = request?.url ?: return false
            return !isAllowedHost(target)
        }

        override fun onPageStarted(view: WebView?, u: String?, favicon: Bitmap?) {
            progress = STARTED_PROGRESS
            loadError = null
        }

        override fun onPageFinished(view: WebView?, u: String?) {
            progress = 1f
        }

        @SuppressLint("WebViewClientOnReceivedSslError")
        override fun onReceivedSslError(view: WebView?, handler: SslErrorHandler?, error: SslError?) {
            if (BuildConfig.DEBUG) handler?.proceed() else handler?.cancel()
        }

        override fun onReceivedError(view: WebView?, request: WebResourceRequest?, error: WebResourceError?) {
            if (request?.isForMainFrame != true) return
            loadError = error?.description?.toString()?.let { UiText.of(it) } ?: UiText.of(R.string.h5_load_failed)
        }
    }

    private fun isAllowedHost(url: Uri): Boolean {
        val allowedHost = Uri.parse(BuildConfig.H5_BASE_URL).host
        if (allowedHost.isNullOrBlank()) return true
        return url.host.equals(allowedHost, ignoreCase = true)
    }

    private companion object {
        const val BRIDGE_NAME = "NearbyApp"
        const val PERCENT = 100f
        const val STARTED_PROGRESS = 0.05f
    }
}
