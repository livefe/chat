package com.nearby.chat.feature.messages

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PushPin
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.layout.boundsInWindow
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.format.formatConversationTime
import com.nearby.chat.core.format.sanitizeNickname
import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppUnreadBadge
import com.nearby.chat.core.ui.UserAvatar

internal val ConversationAvatarSize = 44.dp
private val PinIconSize = 12.dp

@OptIn(ExperimentalFoundationApi::class)
@Composable
internal fun ConversationRow(
    item: ConversationItem,
    onClick: () -> Unit,
    onLongClick: (Rect) -> Unit,
) {
    val context = LocalContext.current
    val name = sanitizeNickname(item.peerNickname)
    var bounds by remember { mutableStateOf(Rect.Zero) }
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .onGloballyPositioned { bounds = it.boundsInWindow() }
            .combinedClickable(
                onClick = onClick,
                onLongClick = { onLongClick(bounds) },
            )
            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        UserAvatar(url = item.peerAvatar, size = ConversationAvatarSize)
        Column(
            modifier = Modifier
                .weight(1f)
                .padding(start = AppSpacing.M),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    name,
                    fontSize = AppType.Subhead,
                    fontWeight = FontWeight.Medium,
                    color = AppColors.TextPrimary,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f),
                )
                if (item.pinned) {
                    Icon(
                        imageVector = Icons.Filled.PushPin,
                        contentDescription = null,
                        tint = AppColors.TextTertiary,
                        modifier = Modifier
                            .padding(start = 6.dp)
                            .size(PinIconSize),
                    )
                }
                Text(
                    formatConversationTime(context, item.lastActiveAt),
                    fontSize = AppType.Micro,
                    color = AppColors.TextMuted,
                    modifier = Modifier.padding(start = AppSpacing.Xs),
                )
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(top = AppSpacing.Xs),
            ) {
                Text(
                    item.lastContent.ifBlank { stringResource(R.string.private_chat_empty) },
                    fontSize = AppType.Caption,
                    color = AppColors.TextMuted,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f),
                )
                if (item.unreadCount > 0) {
                    AppUnreadBadge(
                        count = item.unreadCount,
                        modifier = Modifier.padding(start = AppSpacing.S),
                    )
                }
            }
        }
    }
}
