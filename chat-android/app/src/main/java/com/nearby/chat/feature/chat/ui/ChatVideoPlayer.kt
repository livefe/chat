package com.nearby.chat.feature.chat.ui

import android.net.Uri
import android.view.Gravity
import android.view.ViewGroup
import android.widget.FrameLayout
import android.widget.MediaController
import android.widget.VideoView
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppSpacing

private const val VIDEO_BOTTOM_MARGIN_RATIO = 0.1f
private val CloseButtonSize = 32.dp

@Composable
internal fun ChatVideoPlayer(
    url: String,
    onDismiss: () -> Unit,
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(
            usePlatformDefaultWidth = false,
            decorFitsSystemWindows = false,
        ),
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            AndroidView(
                factory = { ctx ->
                    FrameLayout(ctx).apply {
                        val video = VideoView(ctx)
                        addView(
                            video,
                            FrameLayout.LayoutParams(
                                ViewGroup.LayoutParams.MATCH_PARENT,
                                ViewGroup.LayoutParams.MATCH_PARENT,
                                Gravity.CENTER,
                            ).apply {
                                bottomMargin = (resources.displayMetrics.heightPixels * VIDEO_BOTTOM_MARGIN_RATIO).toInt()
                            },
                        )
                        val controller = MediaController(ctx)
                        video.setMediaController(controller)
                        controller.setAnchorView(this)
                        video.setVideoURI(Uri.parse(url))
                        video.setOnPreparedListener { mp ->
                            val vw = mp.videoWidth
                            val vh = mp.videoHeight
                            val pw = width.takeIf { it > 0 } ?: resources.displayMetrics.widthPixels
                            val ph = height.takeIf { it > 0 } ?: resources.displayMetrics.heightPixels
                            if (vw > 0 && vh > 0) {
                                val scale = minOf(pw.toFloat() / vw, ph.toFloat() / vh)
                                val w = (vw * scale).toInt().coerceAtLeast(1)
                                val h = (vh * scale).toInt().coerceAtLeast(1)
                                video.layoutParams = FrameLayout.LayoutParams(w, h, Gravity.CENTER).apply {
                                    bottomMargin = (ph * VIDEO_BOTTOM_MARGIN_RATIO).toInt()
                                }
                            }
                            video.start()
                        }
                    }
                },
                onRelease = { layout ->
                    (layout.getChildAt(0) as? VideoView)?.stopPlayback()
                },
                modifier = Modifier.fillMaxSize(),
            )
            Icon(
                Icons.Outlined.Close,
                contentDescription = stringResource(R.string.common_close),
                tint = Color.White,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .statusBarsPadding()
                    .padding(AppSpacing.M)
                    .size(CloseButtonSize)
                    .clickable(onClick = onDismiss)
                    .padding(AppSpacing.Xs),
            )
        }
    }
}
