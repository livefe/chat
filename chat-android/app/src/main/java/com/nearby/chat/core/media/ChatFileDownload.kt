package com.nearby.chat.core.media

import android.app.DownloadManager
import android.content.Context
import android.net.Uri
import android.os.Environment
import android.webkit.MimeTypeMap
import com.nearby.chat.R
import com.nearby.chat.core.chat.chatFileDisplayName
import com.nearby.chat.core.chat.chatMediaDisplayUrl
import com.nearby.chat.core.ui.Toaster
import org.json.JSONObject

object ChatFileDownload {
    fun start(context: Context, toaster: Toaster, contentJson: String) {
        val url = chatMediaDisplayUrl(contentJson)
        if (url.isBlank()) {
            toaster.fail(R.string.download_invalid_url)
            return
        }
        val name = sanitizeFileName(
            chatFileDisplayName(contentJson).ifBlank { context.getString(R.string.chat_file_default_name) },
        )
        val mime = parseMimeType(contentJson).ifBlank { guessMime(name) }
        val appCtx = context.applicationContext
        val dm = appCtx.getSystemService(Context.DOWNLOAD_SERVICE) as? DownloadManager
        if (dm == null) {
            toaster.fail(R.string.download_service_unavailable)
            return
        }
        runCatching {
            val request = DownloadManager.Request(Uri.parse(url))
                .setTitle(name)
                .setDescription(appCtx.getString(R.string.download_in_progress))
                .setNotificationVisibility(
                    DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED,
                )
                .setAllowedOverMetered(true)
                .setAllowedOverRoaming(true)
                .setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, name)
            if (mime.isNotBlank()) {
                request.setMimeType(mime)
            }
            dm.enqueue(request)
            toaster.show(R.string.download_started)
        }.onFailure {
            toaster.fail(it, R.string.download_failed)
        }
    }

    private fun parseMimeType(raw: String): String = runCatching {
        JSONObject(raw.trim()).optString("mimeType").trim()
    }.getOrDefault("")

    private fun guessMime(fileName: String): String {
        val ext = fileName.substringAfterLast('.', "").lowercase()
        if (ext.isBlank()) return "application/octet-stream"
        return MimeTypeMap.getSingleton().getMimeTypeFromExtension(ext)
            ?: when (ext) {
                "pdf" -> "application/pdf"
                "doc" -> "application/msword"
                "docx" -> "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                "xls" -> "application/vnd.ms-excel"
                "xlsx" -> "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                "ppt" -> "application/vnd.ms-powerpoint"
                "pptx" -> "application/vnd.openxmlformats-officedocument.presentationml.presentation"
                "txt" -> "text/plain"
                "csv" -> "text/csv"
                "zip" -> "application/zip"
                "rar" -> "application/vnd.rar"
                "7z" -> "application/x-7z-compressed"
                else -> "application/octet-stream"
            }
    }

    private fun sanitizeFileName(name: String): String {
        val cleaned = name
            .replace(Regex("""[\\/:*?"<>|]"""), "_")
            .trim()
            .take(120)
        return cleaned.ifBlank { "file_${System.currentTimeMillis()}" }
    }
}
