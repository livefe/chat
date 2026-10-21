package com.nearby.chat.data.export

import android.content.ClipData
import android.content.ContentValues
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import androidx.core.content.FileProvider
import com.nearby.chat.R
import com.nearby.chat.core.model.PostSignupItem
import com.nearby.chat.core.model.PostSignupListCounts
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.data.repo.PostRepository
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SignupExportRepository @Inject constructor(
    private val postRepository: PostRepository,
    @ApplicationContext private val context: Context,
) {
    suspend fun export(
        postId: Long,
        postTitle: String,
        status: Int?,
        checkIn: Int?,
        onPage: (needCheckIn: Boolean, counts: PostSignupListCounts) -> Unit,
    ) {
        var needCheckIn = false
        val rows = fetchAll(postId, status, checkIn) { pageNeedCheckIn, counts ->
            needCheckIn = pageNeedCheckIn
            onPage(pageNeedCheckIn, counts)
        }
        if (rows.isEmpty()) throw EmptySignupExportException()
        val title = postTitle.ifBlank { postId.toString() }
        val filename = sanitizeFilename(
            context.getString(R.string.signup_export_filename, title, filenameStamp()),
        )
        val bytes = withContext(Dispatchers.Default) { buildWorkbook(rows, needCheckIn) }
        withContext(Dispatchers.IO) { saveOrShare(filename, bytes) }
    }

    private suspend fun fetchAll(
        postId: Long,
        status: Int?,
        checkIn: Int?,
        onPage: (Boolean, PostSignupListCounts) -> Unit,
    ): List<PostSignupItem> {
        val all = mutableListOf<PostSignupItem>()
        var pageNo = 1
        while (true) {
            val data = postRepository.signupList(
                postId = postId,
                page = pageNo,
                pageSize = EXPORT_PAGE_SIZE,
                status = status,
                checkIn = checkIn,
            )
            onPage(data.needCheckIn, data.counts)
            all += data.list
            if (all.size >= data.total || data.list.size < EXPORT_PAGE_SIZE) break
            pageNo += 1
        }
        return all
    }

    private fun buildWorkbook(rows: List<PostSignupItem>, needCheckIn: Boolean): ByteArray {
        val headers = if (needCheckIn) {
            listOf(
                R.string.signup_export_col_nickname,
                R.string.signup_export_col_status,
                R.string.signup_export_col_signup_at,
                R.string.signup_export_col_check_in,
                R.string.signup_export_col_check_in_at,
                R.string.signup_export_col_cancelled_at,
                R.string.signup_export_col_remark,
            )
        } else {
            listOf(
                R.string.signup_export_col_nickname,
                R.string.signup_export_col_status,
                R.string.signup_export_col_signup_at,
                R.string.signup_export_col_cancelled_at,
                R.string.signup_export_col_remark,
            )
        }.map { context.getString(it) }
        val dataRows = rows.map { item ->
            buildList {
                add(item.userNickname)
                add(item.signupStatusText)
                add(formatExportDateTime(item.createdAt))
                if (needCheckIn) {
                    add(
                        if (item.signupStatus == 1) {
                            context.getString(
                                PostSignupRules.checkInLabel(PostSignupRules.hasTimestamp(item.checkedInAt)),
                            )
                        } else {
                            ""
                        },
                    )
                    add(
                        if (PostSignupRules.hasTimestamp(item.checkedInAt)) {
                            formatExportDateTime(item.checkedInAt)
                        } else {
                            ""
                        },
                    )
                }
                add(
                    if (item.cancelledAt.isNotBlank()) {
                        formatExportDateTime(item.cancelledAt)
                    } else {
                        ""
                    },
                )
                add(item.remark.trim())
            }
        }
        return XlsxWriter.build(context.getString(R.string.signup_export_sheet_name), headers, dataRows)
    }

    private fun filenameStamp(now: Date = Date()): String = SimpleDateFormat("yyyyMMdd-HHmmss", Locale.getDefault()).format(now)

    private fun sanitizeFilename(raw: String): String {
        val cleaned = raw.replace(Regex("""[\\/:*?"<>|]"""), "_").trim()
            .removeSuffix(".csv").removeSuffix(".xls").removeSuffix(".xlsx")
            .removeSuffix(".CSV").removeSuffix(".XLS").removeSuffix(".XLSX")
        return "$cleaned.xlsx"
    }

    private fun saveOrShare(filename: String, bytes: ByteArray) {
        val uri = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            saveToDownloads(filename, bytes)
        } else {
            saveToCache(filename, bytes)
        }
        openShareChooser(uri, filename)
    }

    private fun openShareChooser(uri: Uri, filename: String) {
        val viewIntent = Intent(Intent.ACTION_VIEW).apply {
            setDataAndType(uri, XlsxWriter.MIME_XLSX)
            clipData = ClipData.newUri(context.contentResolver, filename, uri)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
        }
        val shareIntent = Intent(Intent.ACTION_SEND).apply {
            type = XlsxWriter.MIME_XLSX
            putExtra(Intent.EXTRA_STREAM, uri)
            putExtra(Intent.EXTRA_SUBJECT, filename)
            clipData = ClipData.newUri(context.contentResolver, filename, uri)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
        }
        val chooserTitle = context.getString(R.string.signup_export_chooser_title, filename)
        val chooser = Intent.createChooser(shareIntent, chooserTitle).apply {
            putExtra(Intent.EXTRA_INITIAL_INTENTS, arrayOf(viewIntent))
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        context.startActivity(chooser)
    }

    private fun saveToCache(filename: String, bytes: ByteArray): Uri {
        val dir = File(context.cacheDir, "share").apply { mkdirs() }
        val file = File(dir, filename)
        file.writeBytes(bytes)
        return FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file,
        )
    }

    private fun saveToDownloads(filename: String, bytes: ByteArray): Uri {
        val values = ContentValues().apply {
            put(MediaStore.Downloads.DISPLAY_NAME, filename)
            put(MediaStore.Downloads.MIME_TYPE, XlsxWriter.MIME_XLSX)
            put(MediaStore.Downloads.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
            put(MediaStore.Downloads.IS_PENDING, 1)
        }
        val resolver = context.contentResolver
        val collection = MediaStore.Downloads.getContentUri(MediaStore.VOLUME_EXTERNAL_PRIMARY)
        val uri = resolver.insert(collection, values)
            ?: error(context.getString(R.string.signup_export_create_failed))
        try {
            resolver.openOutputStream(uri)?.use { it.write(bytes) }
                ?: error(context.getString(R.string.signup_export_write_failed))
            values.clear()
            values.put(MediaStore.Downloads.IS_PENDING, 0)
            resolver.update(uri, values, null, null)
            return uri
        } catch (e: Exception) {
            runCatching { resolver.delete(uri, null, null) }
            throw e
        }
    }

    private fun formatExportDateTime(raw: String): String {
        val value = raw.trim()
        if (value.isEmpty()) return ""
        val m = API_DATE_TIME.matchEntire(value) ?: return value
        val (y, mo, d, hh, mm, ss) = m.destructured
        return context.getString(R.string.signup_export_datetime, y, mo, d, hh, mm, ss)
    }

    private companion object {
        const val EXPORT_PAGE_SIZE = 100
        val API_DATE_TIME = Regex("""^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})$""")
    }
}
