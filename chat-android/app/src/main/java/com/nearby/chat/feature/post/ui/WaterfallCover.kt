package com.nearby.chat.feature.post.ui

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.theme.AppColors

internal const val WATERFALL_RATIO_DEFAULT = 3f / 4f
private const val WATERFALL_RATIO_MIN = 0.66f
private const val WATERFALL_RATIO_MAX = 3f / 4f
private val CoverGradient = listOf(AppColors.NeutralContainer, AppColors.NeutralLine)
internal val WaterfallPreviewGradient = listOf(AppColors.NeutralContainer, AppColors.BorderSoft, AppColors.NeutralShade)
internal val WaterfallPreviewText = AppColors.TextDeep

@Composable
internal fun WaterfallCover(cover: String, label: String) {
    var coverAspect by remember(cover) { mutableFloatStateOf(WATERFALL_RATIO_DEFAULT) }
    var coverReady by remember(cover) { mutableStateOf(false) }
    val coverAlpha by animateFloatAsState(
        targetValue = if (coverReady) 1f else 0f,
        animationSpec = tween(220),
        label = label,
    )
    val context = LocalContext.current
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .aspectRatio(coverAspect)
            .background(Brush.verticalGradient(CoverGradient)),
    ) {
        AsyncImage(
            model = ImageRequest.Builder(context)
                .data(MediaUrl.resolve(cover))
                .crossfade(false)
                .allowHardware(true)
                .build(),
            contentDescription = null,
            contentScale = ContentScale.Crop,
            onSuccess = { state ->
                val size = state.painter.intrinsicSize
                if (size.width > 0f && size.height > 0f) {
                    coverAspect = (size.width / size.height)
                        .coerceIn(WATERFALL_RATIO_MIN, WATERFALL_RATIO_MAX)
                }
                coverReady = true
            },
            modifier = Modifier
                .fillMaxSize()
                .graphicsLayer { alpha = coverAlpha },
        )
    }
}
