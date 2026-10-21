package com.nearby.chat.feature.post.ui

import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes

private val TrackWidth = 40.dp
private val TrackHeight = 22.dp
private val ThumbSize = 18.dp
private val ThumbPadding = 2.dp

@Composable
fun PostSwitch(
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    val offset by animateDpAsState(
        targetValue = if (checked) TrackWidth - ThumbSize - ThumbPadding else ThumbPadding,
        animationSpec = tween(180),
        label = "switchOffset",
    )
    Box(
        modifier = Modifier
            .width(TrackWidth)
            .height(TrackHeight)
            .clip(AppShapes.Capsule)
            .background(if (checked) AppColors.Brand else AppColors.SwitchOff)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = { onCheckedChange(!checked) },
            ),
    ) {
        Box(
            modifier = Modifier
                .offset(x = offset)
                .align(Alignment.CenterStart)
                .size(ThumbSize)
                .shadow(1.dp, CircleShape, clip = false)
                .clip(CircleShape)
                .background(Color.White),
        )
    }
}
