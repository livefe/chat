package com.nearby.chat.feature.forward

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.feature.chat.ui.RoomPlaceTypeColors

internal val ForwardProgressSize = 24.dp
internal val ForwardProgressStroke = 2.dp

@Composable
internal fun ForwardRoomsPane(
    state: ForwardUiState,
    rooms: List<NearbyRoomItem>,
    onSelect: (NearbyRoomItem) -> Unit,
) {
    when {
        state.roomsLoading && state.rooms.isEmpty() -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(
                    color = AppColors.Brand,
                    strokeWidth = ForwardProgressStroke,
                    modifier = Modifier.size(ForwardProgressSize),
                )
            }
        }

        !state.locationReady -> EmptyHint(stringResource(R.string.forward_choose_location), Modifier.fillMaxSize())

        state.rooms.isEmpty() -> EmptyHint(stringResource(R.string.nearby_rooms_empty), Modifier.fillMaxSize())

        rooms.isEmpty() -> EmptyHint(stringResource(R.string.forward_rooms_no_match), Modifier.fillMaxSize())

        else -> {
            LazyColumn(modifier = Modifier.fillMaxSize()) {
                itemsIndexed(rooms, key = { _, it -> it.id }) { index, room ->
                    ForwardListRow(
                        title = room.name.ifBlank { stringResource(R.string.room_default_name) },
                        enabled = !state.sending,
                        onClick = { onSelect(room) },
                    ) {
                        RoomTypeMark(room.placeType)
                    }
                    if (index < rooms.lastIndex) ForwardListDivider()
                }
            }
        }
    }
}

@Composable
private fun RoomTypeMark(placeType: Int) {
    val colors = RoomPlaceTypeColors.of(placeType)
    Box(
        modifier = Modifier
            .size(ForwardLeadingSize)
            .clip(AppShapes.Avatar)
            .background(colors.bg),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = stringResource(RoomPlaceType.option(placeType).mark),
            color = colors.fg,
            fontSize = AppType.Subhead,
            fontWeight = FontWeight.SemiBold,
        )
    }
}
