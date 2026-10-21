package com.nearby.chat.data.repo

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.encodeChatVoice
import com.nearby.chat.core.model.UploadedChatMedia
import com.nearby.chat.core.model.UploadedMedia
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.UploadApi
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody
import okio.BufferedSink
import okio.source
import java.io.File
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UploadRepository @Inject constructor(
    private val uploadApi: UploadApi,
    @ApplicationContext private val context: Context,
) {
    suspend fun uploadImage(uri: Uri): UploadedMedia = uploadFile(
        uri = uri,
        kind = UploadKind.IMAGE,
        maxBytes = IMAGE_MAX_BYTES,
    )

    suspend fun uploadChatMedia(uri: Uri): UploadedChatMedia = withContext(Dispatchers.IO) {
        if (isVideoUri(uri)) {
            val media = uploadVideo(uri)
            UploadedChatMedia(media.id, media.url, MessageContentTypes.VIDEO)
        } else {
            val media = uploadImage(uri)
            UploadedChatMedia(media.id, media.url, MessageContentTypes.IMAGE)
        }
    }

    suspend fun uploadChatFile(uri: Uri): UploadedChatMedia = withContext(Dispatchers.IO) {
        val name = queryDisplayName(uri).orEmpty()
        val media = uploadFile(
            uri = uri,
            kind = UploadKind.FILE,
            maxBytes = FILE_MAX_BYTES,
        )
        UploadedChatMedia(
            mediaId = media.id,
            url = media.url,
            contentType = MessageContentTypes.FILE,
            fileName = name.ifBlank { "file" },
        )
    }

    suspend fun uploadChatVoice(uri: Uri, durationMs: Long): UploadedChatMedia = withContext(Dispatchers.IO) {
        val media = uploadFile(
            uri = uri,
            kind = UploadKind.AUDIO,
            maxBytes = AUDIO_MAX_BYTES,
        )
        UploadedChatMedia(
            mediaId = media.id,
            url = encodeChatVoice(media.id, durationMs),
            contentType = MessageContentTypes.VOICE,
        )
    }

    private fun isVideoUri(uri: Uri): Boolean {
        val mime = context.contentResolver.getType(uri)?.lowercase().orEmpty()
        if (mime.startsWith("video/")) return true
        val name = (queryDisplayName(uri) ?: uri.lastPathSegment.orEmpty()).lowercase()
        return name.endsWith(".mp4") || name.endsWith(".mov") || name.endsWith(".webm") ||
            name.endsWith(".3gp") || name.endsWith(".m4v")
    }

    private suspend fun uploadVideo(uri: Uri): UploadedMedia = uploadFile(
        uri = uri,
        kind = UploadKind.VIDEO,
        maxBytes = VIDEO_MAX_BYTES,
    )

    private suspend fun uploadFile(
        uri: Uri,
        kind: UploadKind,
        maxBytes: Long,
    ): UploadedMedia = withContext(Dispatchers.IO) {
        val resolver = context.contentResolver
        val label = context.getString(kind.label)
        val size = querySize(uri)
        if (size == 0L) error(context.getString(R.string.upload_empty, label))
        if (size > maxBytes) {
            error(context.getString(R.string.upload_too_large, label, maxBytes / (1024 * 1024)))
        }

        val mime = resolver.getType(uri)?.takeIf { kind.matchesMime(it) } ?: kind.defaultMime
        val filename = queryDisplayName(uri) ?: defaultFilename(mime, kind)
        val mediaType = mime.toMediaTypeOrNull()
        val body = object : RequestBody() {
            override fun contentType(): MediaType? = mediaType
            override fun contentLength(): Long = if (size > 0) size else -1
            override fun writeTo(sink: BufferedSink) {
                resolver.openInputStream(uri)?.use { input ->
                    sink.writeAll(input.source())
                } ?: error(context.getString(R.string.upload_read_failed, label))
            }
        }
        val part = MultipartBody.Part.createFormData("file", filename, body)
        val result = when (kind) {
            UploadKind.VIDEO -> uploadApi.uploadVideo(part).unwrap()
            UploadKind.AUDIO -> uploadApi.uploadAudio(part).unwrap()
            UploadKind.IMAGE -> uploadApi.uploadImage(part).unwrap()
            UploadKind.FILE -> uploadApi.uploadFile(part).unwrap()
        }
        if (result.id <= 0L || result.url.isBlank()) error(context.getString(R.string.upload_failed))
        result.toDomain()
    }

    private fun querySize(uri: Uri): Long {
        val cursor = context.contentResolver.query(
            uri,
            arrayOf(OpenableColumns.SIZE),
            null,
            null,
            null,
        ) ?: return fileSizeFallback(uri)
        cursor.use {
            if (!it.moveToFirst()) return fileSizeFallback(uri)
            val idx = it.getColumnIndex(OpenableColumns.SIZE)
            if (idx < 0) return fileSizeFallback(uri)
            val n = it.getLong(idx)
            return if (n >= 0) n else fileSizeFallback(uri)
        }
    }

    private fun fileSizeFallback(uri: Uri): Long {
        if (uri.scheme == "file") {
            val path = uri.path ?: return -1
            val f = File(path)
            return if (f.exists()) f.length() else -1
        }
        return -1
    }

    private fun queryDisplayName(uri: Uri): String? {
        val cursor = context.contentResolver.query(
            uri,
            arrayOf(OpenableColumns.DISPLAY_NAME),
            null,
            null,
            null,
        ) ?: return uri.lastPathSegment?.takeIf { it.isNotBlank() }
        cursor.use {
            if (!it.moveToFirst()) return uri.lastPathSegment?.takeIf { name -> name.isNotBlank() }
            val idx = it.getColumnIndex(OpenableColumns.DISPLAY_NAME)
            if (idx < 0) return uri.lastPathSegment?.takeIf { name -> name.isNotBlank() }
            return it.getString(idx)?.takeIf { name -> name.isNotBlank() }
                ?: uri.lastPathSegment?.takeIf { name -> name.isNotBlank() }
        }
    }

    private fun defaultFilename(mime: String, kind: UploadKind): String {
        val ext = when {
            mime == "image/png" -> "png"
            mime == "image/webp" -> "webp"
            mime == "image/gif" -> "gif"
            mime == "video/webm" -> "webm"
            mime == "video/quicktime" -> "mov"
            mime == "video/3gpp" -> "3gp"
            mime == "audio/mp4" || mime == "audio/m4a" || mime == "audio/x-m4a" -> "m4a"
            mime == "audio/mpeg" -> "mp3"
            mime == "audio/aac" -> "aac"
            mime == "application/pdf" -> "pdf"
            mime.startsWith("video/") -> "mp4"
            mime.startsWith("audio/") -> "m4a"
            mime.startsWith("image/") -> "jpg"
            else -> "bin"
        }
        return when (kind) {
            UploadKind.VIDEO -> "video.$ext"
            UploadKind.AUDIO -> "voice.$ext"
            UploadKind.IMAGE -> "image.$ext"
            UploadKind.FILE -> "file.$ext"
        }
    }

    private enum class UploadKind(
        @StringRes val label: Int,
        val defaultMime: String,
    ) {
        IMAGE(R.string.upload_kind_image, "image/jpeg"),
        VIDEO(R.string.upload_kind_video, "video/mp4"),
        AUDIO(R.string.upload_kind_audio, "audio/mp4"),
        FILE(R.string.upload_kind_file, "application/octet-stream"),
        ;

        fun matchesMime(mime: String): Boolean = when (this) {
            IMAGE -> mime.startsWith("image/")
            VIDEO -> mime.startsWith("video/")
            AUDIO -> mime.startsWith("audio/")
            FILE -> true
        }
    }

    private companion object {
        const val IMAGE_MAX_BYTES = 10L * 1024 * 1024
        const val VIDEO_MAX_BYTES = 50L * 1024 * 1024
        const val AUDIO_MAX_BYTES = 5L * 1024 * 1024
        const val FILE_MAX_BYTES = 20L * 1024 * 1024
    }
}
