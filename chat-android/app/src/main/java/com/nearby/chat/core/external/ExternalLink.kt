package com.nearby.chat.core.external

import android.app.Activity
import android.content.ActivityNotFoundException
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.browser.customtabs.CustomTabColorSchemeParams
import androidx.browser.customtabs.CustomTabsIntent
import androidx.core.graphics.toColorInt

object ExternalLink {
    private val toolbarColor = "#4C8DFF".toColorInt()

    fun open(context: Context, rawUrl: String): Boolean {
        val uri = parseHttpUri(rawUrl) ?: return false
        return openCustomTab(context, uri) || openExternalViewer(context, uri)
    }

    fun parseHttpUri(raw: String): Uri? {
        val trimmed = raw.trim()
        if (trimmed.isEmpty()) return null
        return runCatching {
            val parsed = Uri.parse(trimmed)
            if (parsed.scheme.isNullOrBlank()) {
                Uri.parse("https://$trimmed")
            } else {
                parsed
            }
        }.getOrNull()?.takeIf {
            val scheme = it.scheme?.lowercase()
            scheme == "http" || scheme == "https"
        }
    }

    private fun openCustomTab(context: Context, uri: Uri): Boolean {
        val colorParams = CustomTabColorSchemeParams.Builder()
            .setToolbarColor(toolbarColor)
            .build()
        val customTabsIntent = CustomTabsIntent.Builder()
            .setDefaultColorSchemeParams(colorParams)
            .setShowTitle(true)
            .setUrlBarHidingEnabled(true)
            .setShareState(CustomTabsIntent.SHARE_STATE_ON)
            .build()
        if (context !is Activity) {
            customTabsIntent.intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        return try {
            customTabsIntent.launchUrl(context, uri)
            true
        } catch (_: ActivityNotFoundException) {
            false
        } catch (_: Exception) {
            false
        }
    }

    private fun openExternalViewer(context: Context, uri: Uri): Boolean = try {
        val intent = Intent(Intent.ACTION_VIEW, uri)
        if (context !is Activity) {
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        context.startActivity(intent)
        true
    } catch (_: ActivityNotFoundException) {
        false
    } catch (_: Exception) {
        false
    }
}
