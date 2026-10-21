package com.nearby.chat.feature.chat.ui

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText
import java.io.File

class PickedMediaException(val text: UiText) : IllegalStateException()

internal fun cachePickedMedia(context: Context, uri: Uri): Uri {
    val resolver = context.contentResolver
    val mime = resolver.getType(uri)?.lowercase().orEmpty()
    val displayName = runCatching {
        resolver.query(uri, arrayOf(OpenableColumns.DISPLAY_NAME), null, null, null)?.use { c ->
            if (c.moveToFirst()) {
                val idx = c.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                if (idx >= 0) c.getString(idx)?.takeIf { it.isNotBlank() } else null
            } else {
                null
            }
        }
    }.getOrNull()
        ?: uri.lastPathSegment?.substringAfterLast('/')
    val nameExt = displayName
        ?.substringAfterLast('.', missingDelimiterValue = "")
        ?.lowercase()
        ?.takeIf { it.isNotBlank() && it.length <= 8 && it.all { ch -> ch.isLetterOrDigit() } }
    val ext = when {
        mime.startsWith("video/") -> when {
            "webm" in mime -> "webm"
            "quicktime" in mime -> "mov"
            "3gpp" in mime -> "3gp"
            else -> "mp4"
        }

        mime.startsWith("image/") -> when {
            "png" in mime -> "png"
            "webp" in mime -> "webp"
            "gif" in mime -> "gif"
            else -> "jpg"
        }

        nameExt != null -> nameExt

        mime == "application/pdf" -> "pdf"

        mime.contains("zip") -> "zip"

        mime.contains("rar") -> "rar"

        mime.contains("7z") -> "7z"

        mime.startsWith("text/") -> "txt"

        else -> "bin"
    }
    val kind = when {
        mime.startsWith("video/") -> R.string.upload_kind_video
        mime.startsWith("image/") -> R.string.upload_kind_image
        else -> R.string.upload_kind_file
    }
    val safeBase = displayName
        ?.substringBeforeLast('.')
        ?.replace(Regex("[^a-zA-Z0-9._\\-\\u4e00-\\u9fff]+"), "_")
        ?.take(40)
        ?.takeIf { it.isNotBlank() }
        ?: "pick"
    val out = File(context.cacheDir, "${safeBase}_${System.currentTimeMillis()}.$ext")
    resolver.openInputStream(uri)?.use { input ->
        out.outputStream().use { input.copyTo(it) }
    } ?: throw PickedMediaException(UiText.of(R.string.upload_read_failed, UiText.of(kind)))
    if (out.length() <= 0L) {
        out.delete()
        throw PickedMediaException(UiText.of(R.string.upload_empty, UiText.of(kind)))
    }
    return Uri.fromFile(out)
}
