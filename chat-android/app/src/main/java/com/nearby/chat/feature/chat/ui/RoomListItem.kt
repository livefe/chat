package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.format.formatDistance
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

private val RoomItemPadding = 14.dp
private val RoomTypeMarkSize = 40.dp
private val RoomItemLineGap = 3.dp

@Composable
fun RoomListItem(
    room: NearbyRoomItem,
    onClick: () -> Unit,
) {
    val type = RoomPlaceType.option(room.placeType)
    val colors = RoomPlaceTypeColors.of(room.placeType)
    val typeLabel = RoomPlaceType.display(room.placeType, room.placeTypeText).asString()
    val distance = formatDistance(room.distance).asString()
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppShapes.Card)
            .background(AppColors.CardBg)
            .clickable(onClick = onClick)
            .padding(RoomItemPadding),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(RoomTypeMarkSize)
                .clip(AppShapes.Avatar)
                .background(colors.bg),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = stringResource(type.mark),
                color = colors.fg,
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.SemiBold,
            )
        }
        Spacer(Modifier.width(AppSpacing.M))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = room.name.ifBlank { stringResource(R.string.room_default_name) },
                color = AppColors.TextPrimary,
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Spacer(Modifier.height(RoomItemLineGap))
            Text(
                text = if (typeLabel.isNotBlank()) {
                    stringResource(R.string.room_list_meta_with_type, typeLabel, distance)
                } else {
                    stringResource(R.string.room_list_meta, distance)
                },
                color = AppColors.TextSecondary,
                fontSize = AppType.Caption,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
            modifier = Modifier.size(AppIconSize.Cell),
        )
    }
}
