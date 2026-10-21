package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.height
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppType

internal val ComposerDanger = AppColors.RecordCancel
private val ComposerDangerBg = AppColors.RecordCancelTint
private val CancelSlop = 24.dp
internal val ComposerControlHeight = 40.dp
internal val ComposerLineHeight = 22.sp

@Composable
internal fun ComposerHoldToTalk(
    enabled: Boolean,
    recording: Boolean,
    canceling: Boolean,
    isRecordingNow: () -> Boolean,
    hasPermission: () -> Boolean,
    onRequestPermission: () -> Unit,
    onStart: () -> Unit,
    onCancelingChange: (Boolean) -> Unit,
    onRelease: (cancel: Boolean) -> Unit,
    modifier: Modifier = Modifier,
) {
    val cancelSlopPx = with(LocalDensity.current) { CancelSlop.toPx() }
    Box(
        modifier = modifier
            .height(ComposerControlHeight)
            .clip(AppShapes.Composer)
            .background(
                when {
                    recording && canceling -> ComposerDangerBg
                    recording -> AppColors.Brand.copy(alpha = 0.12f)
                    else -> AppColors.ComposerBg
                },
            )
            .pointerInput(enabled) {
                awaitEachGesture {
                    val down = awaitFirstDown(requireUnconsumed = false)
                    if (!enabled) return@awaitEachGesture
                    if (!hasPermission()) {
                        onRequestPermission()
                        return@awaitEachGesture
                    }
                    val startY = down.position.y
                    onStart()
                    var released = false
                    var shouldCancel = false
                    try {
                        while (true) {
                            val event = awaitPointerEvent()
                            val change = event.changes.firstOrNull { it.id == down.id }
                                ?: event.changes.firstOrNull()
                                ?: break
                            shouldCancel = change.position.y < startY - cancelSlopPx
                            onCancelingChange(shouldCancel)
                            if (!change.pressed) {
                                released = true
                                break
                            }
                            change.consume()
                        }
                    } finally {
                        if (isRecordingNow()) onRelease(shouldCancel || !released)
                    }
                }
            },
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = when {
                recording && canceling -> stringResource(R.string.voice_release_cancel)
                recording -> stringResource(R.string.voice_release_end)
                else -> stringResource(R.string.voice_hold_to_talk)
            },
            color = if (recording && canceling) ComposerDanger else AppColors.TextPrimary,
            fontSize = AppType.Body,
            lineHeight = ComposerLineHeight,
        )
    }
}
