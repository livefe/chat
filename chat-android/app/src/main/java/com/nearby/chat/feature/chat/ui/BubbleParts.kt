package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

internal val BubbleOtherBg = AppColors.Surface
internal val BubbleFailedBg = AppColors.TextMuted
internal const val SENDING_ALPHA = 0.72f
private val BubbleTailWidth = 6.dp
private val BubbleTailHeight = 10.dp
private val BubbleTailTop = 13.5.dp

@Composable
internal fun BubbleTail(
    outgoing: Boolean,
    color: Color,
    modifier: Modifier = Modifier,
) {
    Canvas(
        modifier = modifier
            .offset(y = BubbleTailTop)
            .size(width = BubbleTailWidth, height = BubbleTailHeight),
    ) {
        val path = Path().apply {
            if (outgoing) {
                moveTo(0f, 0f)
                lineTo(size.width, size.height / 2f)
                lineTo(0f, size.height)
            } else {
                moveTo(size.width, 0f)
                lineTo(0f, size.height / 2f)
                lineTo(size.width, size.height)
            }
            close()
        }
        drawPath(path, color)
    }
}

@Composable
internal fun SendFailedLabel(
    color: Color = AppColors.TextSecondary,
    modifier: Modifier = Modifier.padding(top = AppSpacing.Xs, start = AppSpacing.Xxs),
) {
    Text(
        text = stringResource(R.string.chat_send_failed),
        color = color,
        fontSize = AppType.Micro,
        modifier = modifier,
    )
}
