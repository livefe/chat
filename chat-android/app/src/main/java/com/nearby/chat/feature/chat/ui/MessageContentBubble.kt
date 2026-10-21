package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLocationContent
import com.nearby.chat.core.chat.ChatPostContent
import com.nearby.chat.core.chat.chatFileDisplayName
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val TextBubbleHPad = 10.dp
private val TextBubbleVPad = 6.dp
private val TextBubbleLineHeight = 22.sp

@OptIn(ExperimentalFoundationApi::class)
@Composable
internal fun MessageContentBubble(
    content: String,
    isMine: Boolean,
    failed: Boolean,
    location: ChatLocationContent?,
    post: ChatPostContent?,
    isFile: Boolean,
    onClick: (() -> Unit)?,
    onLongClick: (() -> Unit)?,
) {
    val isCard = location != null || post != null || isFile
    val bubbleShape = if (isMine) AppShapes.BubbleOutgoing else AppShapes.BubbleIncoming
    val bubbleSolid = when {
        isCard -> BubbleOtherBg
        isMine && failed -> BubbleFailedBg
        isMine -> AppColors.Brand
        else -> BubbleOtherBg
    }
    Row(verticalAlignment = Alignment.Top) {
        if (!isMine) {
            BubbleTail(outgoing = false, color = bubbleSolid, modifier = Modifier.offset(x = 1.dp))
        }
        Box(
            modifier = Modifier
                .clip(bubbleShape)
                .then(
                    if (isMine && !isCard && !failed) {
                        Modifier.background(AppColors.SendGradient)
                    } else {
                        Modifier.background(bubbleSolid)
                    },
                )
                .then(
                    if (onClick != null || onLongClick != null) {
                        Modifier.combinedClickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = { onClick?.invoke() },
                            onLongClick = onLongClick,
                        )
                    } else {
                        Modifier
                    },
                )
                .then(
                    if (isCard) Modifier else Modifier.padding(horizontal = TextBubbleHPad, vertical = TextBubbleVPad),
                ),
        ) {
            Column {
                when {
                    post != null -> PostCardContent(post)

                    location != null -> LocationCardContent(location)

                    isFile -> FileCardContent(
                        chatFileDisplayName(content).ifBlank { stringResource(R.string.chat_file_default_name) },
                    )

                    else -> Text(
                        text = content,
                        color = if (isMine) Color.White else AppColors.TextPrimary,
                        fontSize = AppType.Body,
                        lineHeight = TextBubbleLineHeight,
                    )
                }
                if (failed) {
                    SendFailedLabel(
                        color = if (isCard || !isMine) AppColors.TextSecondary else Color.White.copy(alpha = 0.9f),
                        modifier = Modifier.padding(top = AppSpacing.Xs, start = if (isCard) AppSpacing.Xs else 0.dp),
                    )
                }
            }
        }
        if (isMine) {
            BubbleTail(outgoing = true, color = bubbleSolid, modifier = Modifier.offset(x = (-1).dp))
        }
    }
}
