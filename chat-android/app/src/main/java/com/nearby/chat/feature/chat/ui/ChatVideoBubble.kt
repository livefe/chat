package com.nearby.chat.feature.chat.ui

import android.content.Context
import android.media.MediaMetadataRetriever
import android.net.Uri
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.sizeIn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.PlayCircle
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.LineHeightStyle
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

private val VideoMaxSize = 220.dp
private val VideoPlayIconSize = 48.dp

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun ChatVideoBubble(
    url: String,
    sending: Boolean,
    failed: Boolean,
    onLongClick: (() -> Unit)?,
    modifier: Modifier = Modifier,
) {
    var playing by remember { mutableStateOf(false) }
    var aspect by remember(url) { mutableFloatStateOf(0f) }
    var durationText by remember(url) { mutableStateOf("") }
    val context = LocalContext.current
    val cover = remember(url, context) {
        if (url.isBlank()) null else videoCoverRequest(context, url)
    }
    LaunchedEffect(url) {
        durationText = ""
        if (url.isBlank()) return@LaunchedEffect
        val ms = withContext(Dispatchers.IO) { readVideoDurationMs(context, url) } ?: return@LaunchedEffect
        durationText = formatVideoDuration(ms)
    }
    Box(
        modifier = modifier.then(
            if (url.isNotBlank() || onLongClick != null) {
                Modifier.combinedClickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = { if (url.isNotBlank()) playing = true },
                    onLongClick = onLongClick,
                )
            } else {
                Modifier
            },
        ),
        contentAlignment = Alignment.Center,
    ) {
        if (cover != null) {
            AsyncImage(
                model = cover,
                contentDescription = stringResource(R.string.upload_kind_video),
                contentScale = ContentScale.Fit,
                onSuccess = { state ->
                    val size = state.painter.intrinsicSize
                    if (size.width > 0f && size.height > 0f) {
                        aspect = size.width / size.height
                    }
                },
                modifier = Modifier
                    .sizeIn(maxWidth = VideoMaxSize, maxHeight = VideoMaxSize)
                    .then(if (aspect > 0f) Modifier.aspectRatio(aspect) else Modifier)
                    .clip(AppShapes.Bubble)
                    .background(AppColors.PageBg),
            )
        }
        Icon(
            Icons.Outlined.PlayCircle,
            contentDescription = stringResource(R.string.video_play),
            tint = Color.White.copy(alpha = if (sending) 0.55f else 0.92f),
            modifier = Modifier
                .align(Alignment.Center)
                .size(VideoPlayIconSize),
        )
        if (durationText.isNotBlank()) {
            Text(
                text = durationText,
                color = Color.White,
                fontSize = AppType.Micro,
                style = TextStyle(
                    lineHeight = AppType.Micro,
                    lineHeightStyle = LineHeightStyle(
                        alignment = LineHeightStyle.Alignment.Center,
                        trim = LineHeightStyle.Trim.Both,
                    ),
                ),
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(AppSpacing.S),
            )
        }
        if (failed) {
            Text(
                text = stringResource(R.string.chat_send_failed),
                color = Color.White.copy(alpha = 0.9f),
                fontSize = AppType.Caption,
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(10.dp),
            )
        }
    }
    if (playing && url.isNotBlank()) {
        ChatVideoPlayer(url = url, onDismiss = { playing = false })
    }
}

private fun formatVideoDuration(ms: Long): String {
    if (ms <= 0L) return ""
    val total = (ms / 1000L).toInt()
    val h = total / 3600
    val m = (total % 3600) / 60
    val s = total % 60
    return if (h > 0) "%d:%02d:%02d".format(h, m, s) else "%d:%02d".format(m, s)
}

private fun readVideoDurationMs(context: Context, url: String): Long? {
    val retriever = MediaMetadataRetriever()
    return try {
        val uri = Uri.parse(url)
        when (uri.scheme?.lowercase()) {
            "http", "https" -> retriever.setDataSource(url, emptyMap())
            else -> retriever.setDataSource(context, uri)
        }
        retriever.extractMetadata(MediaMetadataRetriever.METADATA_KEY_DURATION)?.toLongOrNull()
    } catch (_: Exception) {
        null
    } finally {
        retriever.release()
    }
}
