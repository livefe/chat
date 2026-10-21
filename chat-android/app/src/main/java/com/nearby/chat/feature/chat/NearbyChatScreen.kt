package com.nearby.chat.feature.chat

import android.Manifest
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.rememberMultiplePermissionsState
import com.nearby.chat.R
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.feature.chat.ui.NearbyRoomMarquee
import com.nearby.chat.feature.chat.ui.scrollToLatestIfFollowing

private const val ROOM_MARQUEE_LIMIT = 5

@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun NearbyChatScreen(
    onOpenRoom: (Long) -> Unit,
    onOpenNearbyRooms: () -> Unit,
    onOpenUser: (Long) -> Unit,
    onOpenLocationPick: () -> Unit,
    onOpenShareLocation: () -> Unit,
    onOpenPost: (Long) -> Unit = {},
    onReportMessage: (MessageItem) -> Unit = {},
    onOpenForward: (MessageItem) -> Unit = {},
    viewModel: NearbyChatViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val thread by viewModel.threadState.collectAsStateWithLifecycle()
    val rooms by viewModel.rooms.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    var locationSheet by remember { mutableStateOf(false) }
    val locationPermission = rememberMultiplePermissionsState(
        listOf(
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
        ),
    )

    LaunchedEffect(Unit) {
        viewModel.loadPreferCached(canUseGps = locationPermission.allPermissionsGranted)
        if (!locationPermission.allPermissionsGranted) {
            locationPermission.launchMultiplePermissionRequest()
        }
    }

    LaunchedEffect(locationPermission.allPermissionsGranted) {
        if (locationPermission.allPermissionsGranted && !state.locationReady && !state.locating) {
            viewModel.loadPreferCached(canUseGps = true)
        }
    }

    LaunchedEffect(thread.messages.lastOrNull()?.id, thread.messages.lastOrNull()?.clientMsgId) {
        listState.scrollToLatestIfFollowing(
            last = thread.messages.lastOrNull(),
            selfId = thread.selfId,
            reverseLayout = true,
            latestIndex = 0,
        )
    }

    val chooseLocation = stringResource(R.string.location_pick_title)
    val nearbyLabel = stringResource(R.string.place_nearby)
    val locationLabel = when {
        state.locating -> stringResource(R.string.location_locating)

        state.locationReady -> PlaceNames.display(
            state.placeName,
            isCurrentGps = !state.isBrowsing,
            nearbyLabel = nearbyLabel,
        ).ifEmpty { chooseLocation }

        state.locationFailed -> stringResource(R.string.location_locate_failed)

        else -> chooseLocation
    }

    val marqueeRooms = remember(rooms) { rooms.sortedBy { it.distance }.take(ROOM_MARQUEE_LIMIT) }
    val marqueeKey = remember(marqueeRooms) { marqueeRooms.joinToString(",") { it.id.toString() } }
    var showMarquee by remember { mutableStateOf(false) }
    LaunchedEffect(marqueeKey, state.locationReady) {
        showMarquee = state.locationReady && marqueeRooms.isNotEmpty()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.PageBg)
            .statusBarsPadding(),
    ) {
        NearbyChatHeader(
            locationLabel = locationLabel,
            locationReady = state.locationReady,
            onlineCount = state.onlineCount,
            onLocationClick = { locationSheet = true },
            onOpenNearbyRooms = onOpenNearbyRooms,
        )

        Box(Modifier.weight(1f).fillMaxWidth()) {
            NearbyChatMessagePane(
                loading = state.loading,
                locationReady = state.locationReady,
                chat = viewModel.thread,
                threadState = thread,
                listState = listState,
                onPickLocation = onOpenShareLocation,
                onLoadMore = viewModel::loadMoreMessages,
                onOpenUser = onOpenUser,
                onOpenPost = onOpenPost,
                onForward = onOpenForward,
                onReportMessage = onReportMessage,
            )

            if (showMarquee) {
                NearbyRoomMarquee(
                    rooms = marqueeRooms,
                    onRoomClick = { onOpenRoom(it.id) },
                    onFinished = { showMarquee = false },
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .fillMaxWidth()
                        .padding(top = AppSpacing.S),
                )
            }
        }
    }

    if (locationSheet) {
        NearbyLocationSheet(
            gpsSelected = state.locationReady && !state.isBrowsing,
            mapSelected = state.locationReady && state.isBrowsing,
            onDismiss = { locationSheet = false },
            onRelocate = {
                locationSheet = false
                if (locationPermission.allPermissionsGranted) {
                    viewModel.refreshGpsAndBootstrap()
                } else {
                    locationPermission.launchMultiplePermissionRequest()
                }
            },
            onMapPick = {
                locationSheet = false
                onOpenLocationPick()
            },
        )
    }
}
