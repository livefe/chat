package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
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
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val ChipShape = RoundedCornerShape(6.dp)
private val ChipMaxWidth = 120.dp

@Composable
internal fun NearbyRoomSuggestions(
    rooms: List<NearbyRoomItem>,
    selectedRoomId: Long,
    loading: Boolean,
    onSelect: (NearbyRoomItem) -> Unit,
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = AppSpacing.L, end = AppSpacing.L, bottom = AppSpacing.M),
    ) {
        Text(
            stringResource(R.string.place_nearby),
            fontSize = AppType.Caption,
            color = AppColors.TextTertiary,
            modifier = Modifier.padding(end = 10.dp),
        )
        Row(
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
            modifier = Modifier
                .weight(1f)
                .horizontalScroll(rememberScrollState()),
        ) {
            if (loading && rooms.isEmpty()) {
                Text(
                    stringResource(R.string.post_nearby_searching),
                    fontSize = AppType.Caption,
                    color = AppColors.TextTertiary,
                    modifier = Modifier.align(Alignment.CenterVertically),
                )
            }
            val defaultName = stringResource(R.string.room_default_name)
            rooms.forEach { room ->
                val selected = selectedRoomId == room.id
                Text(
                    room.name.ifBlank { defaultName },
                    fontSize = AppType.Caption,
                    fontWeight = if (selected) FontWeight.Medium else FontWeight.Normal,
                    color = if (selected) AppColors.TextPrimary else AppColors.TextTertiary,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier
                        .widthIn(max = ChipMaxWidth)
                        .clip(ChipShape)
                        .background(if (selected) AppColors.BrandContainer else AppColors.PageBg)
                        .clickable { onSelect(room) }
                        .padding(horizontal = 10.dp, vertical = 5.dp),
                )
            }
        }
    }
}
