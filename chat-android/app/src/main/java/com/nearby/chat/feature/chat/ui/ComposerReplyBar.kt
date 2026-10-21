package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.displayName
import com.nearby.chat.core.chat.openChatLocationMap
import com.nearby.chat.core.chat.parseChatLocation
import com.nearby.chat.core.chat.replyQuoteText
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

private val ReplyBarShape = RoundedCornerShape(4.dp)
private val ReplyBarBg = AppColors.OverlayMuted
private val ReplyCloseSize = 24.dp
private val ReplyBarVPad = 6.dp

@Composable
internal fun ComposerReplyBar(
    replyTo: MessageItem,
    onClear: () -> Unit,
) {
    val context = LocalContext.current
    val location = if (replyTo.contentType == MessageContentTypes.LOCATION) {
        parseChatLocation(replyTo.content)
    } else {
        null
    }
    val label = location?.displayName() ?: replyQuoteText(replyTo.contentType, replyTo.content).asString()
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.M, vertical = AppSpacing.Xs)
            .clip(ReplyBarShape)
            .background(ReplyBarBg)
            .padding(horizontal = AppSpacing.S, vertical = ReplyBarVPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = stringResource(R.string.chat_reply_prefix, replyTo.fromName),
            fontSize = AppType.Caption,
            color = AppColors.TextMuted,
            maxLines = 1,
        )
        Text(
            text = label,
            fontSize = AppType.Caption,
            color = AppColors.TextMuted,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier
                .weight(1f)
                .then(
                    if (location != null) {
                        Modifier.clickable { openChatLocationMap(context, location) }
                    } else {
                        Modifier
                    },
                ),
        )
        Icon(
            Icons.Outlined.Close,
            contentDescription = stringResource(R.string.chat_reply_cancel),
            tint = AppColors.TextTertiary,
            modifier = Modifier
                .size(ReplyCloseSize)
                .clickable(onClick = onClear)
                .padding(AppSpacing.Xs),
        )
    }
}
