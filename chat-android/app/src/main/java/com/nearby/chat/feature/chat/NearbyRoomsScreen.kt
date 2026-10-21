package com.nearby.chat.feature.chat

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPullToRefreshBox
import com.nearby.chat.core.ui.AppTopBarTitle
import com.nearby.chat.core.ui.EmptyHint
import com.nearby.chat.core.ui.appTopAppBarColors
import com.nearby.chat.feature.chat.ui.RoomListItem

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NearbyRoomsScreen(
    onBack: () -> Unit,
    onOpenRoom: (Long) -> Unit,
    onOpenRoomCreate: () -> Unit,
    viewModel: NearbyRoomsViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val rooms = state.rooms

    Scaffold(
        containerColor = AppColors.PageBg,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = {
            TopAppBar(
                title = {
                    AppTopBarTitle(stringResource(R.string.nearby_rooms_title))
                },
                navigationIcon = {
                    AppBackIconButton(onClick = onBack)
                },
                actions = {
                    IconButton(onClick = onOpenRoomCreate) {
                        Icon(
                            Icons.Outlined.Add,
                            contentDescription = stringResource(R.string.nearby_room_create),
                            tint = AppColors.TextPrimary,
                            modifier = Modifier.size(AppIconSize.Cell),
                        )
                    }
                },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        AppPullToRefreshBox(
            isRefreshing = state.refreshing,
            onRefresh = viewModel::refresh,
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            when {
                state.loading && rooms.isEmpty() -> {
                    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                    }
                }

                rooms.isEmpty() -> {
                    EmptyHint(
                        stringResource(
                            if (state.locationReady) R.string.nearby_rooms_empty else R.string.nearby_rooms_need_location,
                        ),
                        Modifier.fillMaxSize(),
                    )
                }

                else -> {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(horizontal = AppSpacing.M, vertical = AppSpacing.S),
                        verticalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    ) {
                        items(rooms, key = { it.id }) { room ->
                            RoomListItem(room = room, onClick = { onOpenRoom(room.id) })
                        }
                    }
                }
            }
        }
    }
}
