package com.nearby.chat.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ChatBubbleOutline
import androidx.compose.material.icons.outlined.KeyboardArrowDown
import androidx.compose.material.icons.outlined.MyLocation
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.dismissKeyboardOnTap

private val HeaderTopPadding = 14.dp
private val OnlineDotSize = 6.dp
private val OnlineDotGap = 5.dp
private val RoomsEntrySize = 36.dp

@Composable
internal fun NearbyChatHeader(
    locationLabel: String,
    locationReady: Boolean,
    onlineCount: Int,
    onLocationClick: () -> Unit,
    onOpenNearbyRooms: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .dismissKeyboardOnTap()
            .padding(start = AppSpacing.L, end = AppSpacing.L, top = HeaderTopPadding, bottom = AppSpacing.Xs),
    ) {
        BoxWithConstraints(modifier = Modifier.fillMaxWidth()) {
            val titleMaxWidth = maxWidth / 2
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Row(
                    modifier = Modifier
                        .widthIn(max = titleMaxWidth)
                        .clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = onLocationClick,
                        ),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Icon(
                        Icons.Outlined.MyLocation,
                        contentDescription = null,
                        tint = AppColors.TextPrimary,
                        modifier = Modifier.size(AppIconSize.Inline),
                    )
                    Spacer(Modifier.width(AppSpacing.S))
                    Text(
                        text = locationLabel,
                        color = AppColors.TextPrimary,
                        fontSize = AppType.PageTitle,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        softWrap = false,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.weight(1f, fill = false),
                    )
                    Icon(
                        Icons.Outlined.KeyboardArrowDown,
                        contentDescription = null,
                        tint = AppColors.TextPrimary,
                        modifier = Modifier.size(AppIconSize.Inline),
                    )
                }
                if (locationReady) {
                    Spacer(Modifier.width(AppSpacing.S))
                    Box(
                        Modifier
                            .size(OnlineDotSize)
                            .clip(CircleShape)
                            .background(AppColors.Success),
                    )
                    Spacer(Modifier.width(OnlineDotGap))
                    Text(
                        text = stringResource(R.string.chat_online_count, onlineCount),
                        color = AppColors.TextMuted,
                        fontSize = AppType.Caption,
                        maxLines = 1,
                        softWrap = false,
                    )
                }
                Spacer(Modifier.weight(1f))
                RoomsEntryButton(onClick = onOpenNearbyRooms)
            }
        }
    }
}

@Composable
private fun RoomsEntryButton(onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(RoomsEntrySize)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            Icons.Outlined.ChatBubbleOutline,
            contentDescription = stringResource(R.string.nearby_rooms_title),
            tint = AppColors.TextPrimary,
            modifier = Modifier.size(AppIconSize.Cell),
        )
    }
}
