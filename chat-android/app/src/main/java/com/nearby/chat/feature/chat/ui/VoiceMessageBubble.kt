package com.nearby.chat.feature.chat.ui

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.chat.formatVoiceDuration
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.LocalToaster
import com.nearby.chat.core.ui.LocalVoicePlayer

private val VoiceBubbleHeight = 40.dp
private val VoiceBarWidth = 3.dp
private val VoiceBarGap = 3.dp
private val VoiceBarsHeight = 14.dp
private val VoiceBarShape = RoundedCornerShape(1.dp)
private const val VOICE_BASE_WIDTH_DP = 72
private const val VOICE_WIDTH_PER_SECOND_DP = 4
private const val VOICE_MAX_WIDTH_DP = 180

@OptIn(ExperimentalFoundationApi::class)
@Composable
internal fun VoiceMessageBubble(
    voiceUrl: String,
    durationMs: Long,
    isMine: Boolean,
    playKey: String,
    sending: Boolean,
    failed: Boolean,
    onLongClick: (() -> Unit)?,
) {
    val voicePlayer = LocalVoicePlayer.current
    val toaster = LocalToaster.current
    val playingKey by voicePlayer.playingKey.collectAsStateWithLifecycle()
    val playing = playingKey.isNotBlank() && playingKey == playKey
    DisposableEffect(playKey) {
        onDispose {
            if (voicePlayer.isPlaying(playKey)) voicePlayer.stop()
        }
    }
    val widthDp = voiceBubbleWidth(durationMs)
    val fg = if (isMine) Color.White else AppColors.TextPrimary
    val bg = when {
        failed && isMine -> BubbleFailedBg
        isMine -> AppColors.Brand
        else -> BubbleOtherBg
    }
    Column(horizontalAlignment = if (isMine) Alignment.End else Alignment.Start) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = if (isMine) Arrangement.End else Arrangement.Start,
        ) {
            if (!isMine) {
                BubbleTail(outgoing = false, color = bg, modifier = Modifier.offset(x = 1.dp))
            }
            Row(
                modifier = Modifier
                    .width(widthDp)
                    .height(VoiceBubbleHeight)
                    .clip(if (isMine) AppShapes.BubbleOutgoing else AppShapes.BubbleIncoming)
                    .then(
                        if (isMine && !failed) {
                            Modifier.background(AppColors.SendGradient)
                        } else {
                            Modifier.background(bg)
                        },
                    )
                    .then(
                        if (onLongClick != null || voiceUrl.isNotBlank()) {
                            Modifier.combinedClickable(
                                interactionSource = remember { MutableInteractionSource() },
                                indication = null,
                                onClick = {
                                    if (voiceUrl.isBlank()) return@combinedClickable
                                    runCatching { voicePlayer.toggle(playKey, voiceUrl) }
                                        .onFailure { toaster.fail(R.string.voice_play_failed) }
                                },
                                onLongClick = onLongClick,
                            )
                        } else {
                            Modifier
                        },
                    )
                    .padding(horizontal = AppSpacing.M)
                    .alpha(if (sending) SENDING_ALPHA else 1f),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(
                    AppSpacing.S,
                    if (isMine) Alignment.End else Alignment.Start,
                ),
            ) {
                if (isMine) {
                    Text(text = formatVoiceDuration(durationMs), color = fg, fontSize = AppType.Secondary)
                    VoiceWaveBars(playing = playing, color = fg)
                } else {
                    VoiceWaveBars(playing = playing, color = fg)
                    Text(text = formatVoiceDuration(durationMs), color = fg, fontSize = AppType.Secondary)
                }
            }
            if (isMine) {
                BubbleTail(outgoing = true, color = bg, modifier = Modifier.offset(x = (-1).dp))
            }
        }
        if (failed) {
            SendFailedLabel()
        }
    }
}

private fun voiceBubbleWidth(durationMs: Long): Dp {
    val seconds = (durationMs / 1000.0).toInt().coerceAtLeast(1)
    return (VOICE_BASE_WIDTH_DP + seconds * VOICE_WIDTH_PER_SECOND_DP).coerceAtMost(VOICE_MAX_WIDTH_DP).dp
}

@Composable
private fun VoiceWaveBars(playing: Boolean, color: Color) {
    val transition = rememberInfiniteTransition(label = "voice-wave")
    val t1 by transition.animateFloat(
        initialValue = 0.55f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(900, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "w1",
    )
    val t2 by transition.animateFloat(
        initialValue = 0.55f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(900, delayMillis = 150, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "w2",
    )
    val t3 by transition.animateFloat(
        initialValue = 0.55f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(900, delayMillis = 300, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse,
        ),
        label = "w3",
    )
    Row(
        horizontalArrangement = Arrangement.spacedBy(VoiceBarGap),
        verticalAlignment = Alignment.Bottom,
        modifier = Modifier.height(VoiceBarsHeight),
    ) {
        VoiceBar(height = 6.dp, scale = if (playing) t1 else 1f, color = color)
        VoiceBar(height = 12.dp, scale = if (playing) t2 else 1f, color = color)
        VoiceBar(height = 8.dp, scale = if (playing) t3 else 1f, color = color)
    }
}

@Composable
private fun VoiceBar(height: Dp, scale: Float, color: Color) {
    Box(
        Modifier
            .width(VoiceBarWidth)
            .height(height)
            .graphicsLayer { scaleY = scale }
            .clip(VoiceBarShape)
            .background(color.copy(alpha = 0.85f)),
    )
}
