package com.nearby.chat.core.share

import android.content.ClipData
import android.content.ClipboardManager
import android.content.ContentValues
import android.content.Context
import android.content.Intent
import android.graphics.Bitmap
import android.graphics.Canvas
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import androidx.core.content.FileProvider
import coil.ImageLoader
import coil.request.ImageRequest
import coil.request.SuccessResult
import com.nearby.chat.R
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.media.QrCodeEncoder
import com.nearby.chat.core.ui.Toaster
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.io.FileOutputStream
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PostShare @Inject constructor(
    @ApplicationContext private val context: Context,
    private val toaster: Toaster,
) {
    @Volatile
    private var sharingInFlight = false

    fun isAppInstalled(packageName: String): Boolean = runCatching {
        context.packageManager.getPackageInfo(packageName, 0)
        true
    }.getOrDefault(false)

    private fun postShareUrl(postId: Long, channel: ShareLinks.Channel): String = ShareLinks.webUrl(ShareLinks.Target.Post(postId), channel)

    fun copyLink(postId: Long) {
        val url = postShareUrl(postId, ShareLinks.Channel.COPY)
        if (url.isBlank()) {
            toaster.fail(R.string.share_link_not_configured)
            return
        }
        val cm = context.getSystemService(ClipboardManager::class.java)
        if (cm == null) {
            toaster.fail(R.string.copy_failed)
            return
        }
        cm.setPrimaryClip(ClipData.newPlainText("share", url))
        toaster.show(R.string.share_link_copied)
    }

    suspend fun saveImage(
        postId: Long,
        title: String,
        content: String,
        imageUrl: String? = null,
    ) {
        if (sharingInFlight) return
        sharingInFlight = true
        try {
            val url = postShareUrl(postId, ShareLinks.Channel.QR)
            if (url.isBlank()) {
                toaster.fail(R.string.share_link_not_configured)
                return
            }
            toaster.show(R.string.share_generating_image)
            val file = runCatching {
                buildShareJpeg(postId, title, content, imageUrl, url)
            }.getOrElse {
                toaster.fail(R.string.share_save_failed_retry)
                return
            }
            runCatching {
                withContext(Dispatchers.IO) {
                    saveJpegToGallery(postId, file)
                }
            }.onSuccess {
                toaster.show(R.string.share_saved_to_gallery)
            }.onFailure {
                toaster.fail(R.string.share_save_failed_retry)
            }
        } finally {
            sharingInFlight = false
        }
    }

    suspend fun share(
        postId: Long,
        title: String,
        content: String,
        imageUrl: String? = null,
        targetPackage: String? = null,
    ) {
        if (sharingInFlight) return
        sharingInFlight = true
        try {
            shareInternal(postId, title, content, imageUrl, targetPackage)
        } finally {
            sharingInFlight = false
        }
    }

    private suspend fun shareInternal(
        postId: Long,
        title: String,
        content: String,
        imageUrl: String?,
        targetPackage: String?,
    ) {
        val channel = when (targetPackage) {
            PKG_WECHAT -> ShareLinks.Channel.WECHAT
            PKG_XHS -> ShareLinks.Channel.XHS
            PKG_DOUYIN -> ShareLinks.Channel.DOUYIN
            else -> ShareLinks.Channel.SYSTEM
        }
        val url = postShareUrl(postId, channel)
        if (url.isBlank()) {
            toaster.fail(R.string.share_link_not_configured)
            return
        }
        val target = targetPackage?.takeIf { it.isNotBlank() }
        val appName = when (target) {
            PKG_WECHAT -> context.getString(R.string.share_app_wechat)
            PKG_XHS -> context.getString(R.string.share_app_xhs)
            PKG_DOUYIN -> context.getString(R.string.share_app_douyin)
            else -> null
        }
        if (target != null && !isAppInstalled(target)) {
            toaster.fail(R.string.share_app_not_installed, appName ?: context.getString(R.string.share_app_generic))
            return
        }
        toaster.show(R.string.share_generating_image)
        val displayTitle = title.ifBlank { context.getString(R.string.chat_post_default_title) }
        val file = runCatching {
            buildShareJpeg(postId, title, content, imageUrl, url)
        }.getOrElse {
            toaster.fail(R.string.share_failed_retry)
            return
        }

        val uri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file,
        )
        val intent = Intent(Intent.ACTION_SEND).apply {
            type = "image/jpeg"
            putExtra(Intent.EXTRA_STREAM, uri)
            putExtra(Intent.EXTRA_SUBJECT, displayTitle)
            putExtra(Intent.EXTRA_TEXT, "$displayTitle\n$url")
            clipData = ClipData.newUri(context.contentResolver, "share", uri)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
            if (target != null) setPackage(target)
        }
        if (target != null) {
            context.grantUriPermission(target, uri, Intent.FLAG_GRANT_READ_URI_PERMISSION)
        }
        runCatching {
            if (target == null) {
                val chooser = Intent.createChooser(intent, context.getString(R.string.forward_share_post_title))
                    .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                context.startActivity(chooser)
            } else {
                context.startActivity(intent)
            }
        }.onFailure {
            if (appName != null) {
                toaster.fail(R.string.share_app_open_failed, appName)
            } else {
                toaster.fail(R.string.share_failed_retry)
            }
        }
    }

    private suspend fun buildShareJpeg(
        postId: Long,
        title: String,
        content: String,
        imageUrl: String?,
        url: String,
    ): File = withContext(Dispatchers.IO) {
        val cover = loadCoverBitmap(imageUrl)
        val qr = QrCodeEncoder.encode(url, ShareCardRenderer.QR_SIZE)
        val card = ShareCardRenderer.render(
            title = title.ifBlank { context.getString(R.string.chat_post_default_title) },
            content = content,
            cover = cover,
            qr = qr,
            appLabel = context.getString(R.string.app_name),
            scanTip = context.getString(R.string.share_scan_tip),
        )
        writeShareJpeg(postId, card).also {
            card.recycle()
            qr.recycle()
        }
    }

    private fun saveJpegToGallery(postId: Long, file: File) {
        val filename = "post_$postId.jpg"
        val bytes = file.readBytes()
        val values = ContentValues().apply {
            put(MediaStore.Images.Media.DISPLAY_NAME, filename)
            put(MediaStore.Images.Media.MIME_TYPE, "image/jpeg")
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                put(MediaStore.Images.Media.RELATIVE_PATH, Environment.DIRECTORY_PICTURES)
                put(MediaStore.Images.Media.IS_PENDING, 1)
            }
        }
        val resolver = context.contentResolver
        val collection = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            MediaStore.Images.Media.getContentUri(MediaStore.VOLUME_EXTERNAL_PRIMARY)
        } else {
            MediaStore.Images.Media.EXTERNAL_CONTENT_URI
        }
        val uri = resolver.insert(collection, values) ?: error("MediaStore insert failed")
        resolver.openOutputStream(uri)?.use { it.write(bytes) } ?: error("MediaStore write failed")
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            values.clear()
            values.put(MediaStore.Images.Media.IS_PENDING, 0)
            resolver.update(uri, values, null, null)
        }
    }

    private suspend fun loadCoverBitmap(imageUrl: String?): Bitmap? {
        val resolved = MediaUrl.resolve(imageUrl)
        if (resolved.isBlank()) return null
        val loader = ImageLoader(context)
        val request = ImageRequest.Builder(context)
            .data(resolved)
            .allowHardware(false)
            .build()
        val result = loader.execute(request)
        if (result !is SuccessResult) return null
        val drawable = result.drawable
        val w = drawable.intrinsicWidth.coerceAtLeast(1)
        val h = drawable.intrinsicHeight.coerceAtLeast(1)
        val bmp = Bitmap.createBitmap(w, h, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bmp)
        drawable.setBounds(0, 0, w, h)
        drawable.draw(canvas)
        return bmp
    }

    private fun writeShareJpeg(postId: Long, card: Bitmap): File {
        val dir = File(context.cacheDir, "share").apply { mkdirs() }
        val file = File(dir, "post_$postId.jpg")
        FileOutputStream(file).use { out ->
            card.compress(Bitmap.CompressFormat.JPEG, 92, out)
        }
        return file
    }

    companion object {
        const val PKG_WECHAT = "com.tencent.mm"
        const val PKG_XHS = "com.xingin.xhs"
        const val PKG_DOUYIN = "com.ss.android.ugc.aweme"
    }
}
