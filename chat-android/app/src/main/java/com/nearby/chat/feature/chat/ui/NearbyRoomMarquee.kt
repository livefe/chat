package com.nearby.chat.feature.chat.ui

import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.waitForUpOrCancellation
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import kotlinx.coroutines.delay

private val RoomMarqueeChipHeight = 37.dp
private val RoomMarqueeSpeedDp = 1.dp
private val RoomMarqueeGap = 14.dp
private val RoomMarqueeChipLineHeight = 21.sp
private val RoomMarqueeShadow = 2.dp
private val RoomMarqueeAmbient = AppColors.NeutralTint
private val RoomMarqueeSpot = AppColors.ShadowFaint
private const val FADE_IN_MS = 360
private const val FADE_OUT_MS = 420
private const val ENTER_DELAY_MS = 520L
private const val FRAME_MS = 16L
private const val EXIT_DELAY_MS = 180L

@Composable
fun NearbyRoomMarquee(
    rooms: List<NearbyRoomItem>,
    onRoomClick: (NearbyRoomItem) -> Unit,
    onFinished: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val listState = rememberLazyListState()
    var paused by remember { mutableStateOf(false) }
    var fading by remember { mutableStateOf(false) }
    var entered by remember { mutableStateOf(false) }

    val alpha by animateFloatAsState(
        targetValue = when {
            fading -> 0f
            entered -> 1f
            else -> 0f
        },
        animationSpec = tween(durationMillis = if (fading) FADE_OUT_MS else FADE_IN_MS, easing = FastOutSlowInEasing),
        label = "marqueeAlpha",
        finishedListener = { value ->
            if (fading && value == 0f) onFinished()
        },
    )

    BoxWithConstraints(
        modifier = modifier
            .height(RoomMarqueeChipHeight)
            .alpha(alpha),
    ) {
        val viewportWidth = maxWidth
        val stepPx = with(LocalDensity.current) { RoomMarqueeSpeedDp.toPx() }

        LaunchedEffect(rooms, viewportWidth, stepPx) {
            if (rooms.isEmpty()) {
                onFinished()
                return@LaunchedEffect
            }
            listState.scrollToItem(0, 0)
            entered = true
            delay(ENTER_DELAY_MS)

            while (listState.canScrollForward) {
                if (!paused) {
                    val consumed = listState.dispatchRawDelta(stepPx)
                    if (consumed == 0f && !listState.canScrollForward) break
                }
                delay(FRAME_MS)
            }
            delay(EXIT_DELAY_MS)
            fading = true
        }

        LazyRow(
            state = listState,
            modifier = Modifier
                .fillMaxSize()
                .pointerInput(Unit) {
                    awaitEachGesture {
                        awaitFirstDown(requireUnconsumed = false)
                        paused = true
                        waitForUpOrCancellation()
                        paused = false
                    }
                },
            contentPadding = PaddingValues(start = viewportWidth, end = viewportWidth),
            horizontalArrangement = Arrangement.spacedBy(RoomMarqueeGap),
            userScrollEnabled = true,
        ) {
            items(rooms, key = { it.id }) { room ->
                RoomMarqueeChip(room = room, onClick = { onRoomClick(room) })
            }
        }
    }
}

@Composable
private fun RoomMarqueeChip(
    room: NearbyRoomItem,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .height(RoomMarqueeChipHeight)
            .shadow(
                elevation = RoomMarqueeShadow,
                shape = AppShapes.Bubble,
                ambientColor = RoomMarqueeAmbient,
                spotColor = RoomMarqueeSpot,
                clip = false,
            )
            .clip(AppShapes.Bubble)
            .background(AppColors.CardBg)
            .clickable(onClick = onClick)
            .padding(horizontal = AppSpacing.M),
        contentAlignment = Alignment.CenterStart,
    ) {
        Text(
            text = room.name.ifBlank { stringResource(R.string.room_default_name) },
            color = AppColors.TextPrimary,
            fontSize = AppType.Secondary,
            lineHeight = RoomMarqueeChipLineHeight,
            fontWeight = FontWeight.Medium,
            maxLines = 1,
            softWrap = false,
        )
    }
}
