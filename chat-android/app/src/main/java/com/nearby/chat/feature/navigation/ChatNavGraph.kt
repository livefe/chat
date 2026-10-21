package com.nearby.chat.feature.navigation

import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.nearby.chat.R
import com.nearby.chat.feature.chat.NearbyRoomsScreen
import com.nearby.chat.feature.chat.RoomApplicationsScreen
import com.nearby.chat.feature.chat.RoomCreateScreen
import com.nearby.chat.feature.chat.RoomCreateViewModel
import com.nearby.chat.feature.chat.RoomDetailScreen
import com.nearby.chat.feature.chat.RoomDetailViewModel
import com.nearby.chat.feature.forward.ForwardScreen
import com.nearby.chat.feature.privatechat.PrivateChatScreen
import com.nearby.chat.feature.privatechat.PrivateChatViewModel
import com.nearby.chat.feature.support.SupportChatScreen

internal fun NavGraphBuilder.chatGraph(navigator: AppNavigator) {
    composable(
        route = Routes.PRIVATE_CHAT,
        arguments = listOf(
            navArgument("peerId") { type = NavType.LongType },
            navArgument("name") {
                type = NavType.StringType
                defaultValue = ""
                nullable = true
            },
        ),
    ) { entry ->
        val messageLabel = stringResource(R.string.report_target_message)
        val viewModel: PrivateChatViewModel = hiltViewModel()
        ConsumeShareLocation(entry) { viewModel.thread.sendLocation(it.lat, it.lng, it.placeName) }
        PrivateChatScreen(
            peerId = entry.arguments?.getLong("peerId") ?: 0L,
            peerName = entry.arguments?.getString("name").orEmpty(),
            onBack = navigator::back,
            onOpenProfile = { navigator.navigate(Routes.userProfile(it)) },
            onOpenPost = { navigator.navigate(Routes.postDetail(it)) },
            onPickLocation = navigator::pickShareLocation,
            onReportMessage = { navigator.reportMessage(it, messageLabel) },
            onOpenForward = navigator::forwardMessage,
            viewModel = viewModel,
        )
    }
    composable(Routes.SUPPORT_CHAT) {
        SupportChatScreen(onBack = navigator::back)
    }
    composable(Routes.NEARBY_ROOMS) {
        NearbyRoomsScreen(
            onBack = navigator::back,
            onOpenRoom = { navigator.navigate(Routes.room(it)) },
            onOpenRoomCreate = { navigator.navigate(Routes.ROOM_CREATE) },
        )
    }
    composable(Routes.FORWARD) {
        ForwardScreen(onBack = navigator::back)
    }
    composable(
        route = Routes.ROOM,
        arguments = listOf(navArgument("roomId") { type = NavType.LongType }),
    ) { entry ->
        val messageLabel = stringResource(R.string.report_target_message)
        val viewModel: RoomDetailViewModel = hiltViewModel()
        ConsumeShareLocation(entry) { viewModel.enqueueShareLocation(it.lat, it.lng, it.placeName) }
        RoomDetailScreen(
            roomId = entry.arguments?.getLong("roomId") ?: 0L,
            onBack = navigator::back,
            onOpenPost = { navigator.navigate(Routes.postDetail(it)) },
            onPickLocation = navigator::pickShareLocation,
            onReportMessage = { navigator.reportMessage(it, messageLabel) },
            onOpenForward = navigator::forwardMessage,
            viewModel = viewModel,
        )
    }
    composable(Routes.ROOM_CREATE) { entry ->
        val viewModel: RoomCreateViewModel = hiltViewModel()
        val handle = entry.savedStateHandle
        val lat = handle.get<Double>(NavResultKeys.PICK_LAT)
        val lng = handle.get<Double>(NavResultKeys.PICK_LNG)
        val place = handle.get<String>(NavResultKeys.PICK_PLACE)
        LaunchedEffect(lat, lng, place) {
            if (lat != null && lng != null) {
                viewModel.applyLocation(lat, lng, place.orEmpty())
                handle.remove<Double>(NavResultKeys.PICK_LAT)
                handle.remove<Double>(NavResultKeys.PICK_LNG)
                handle.remove<String>(NavResultKeys.PICK_PLACE)
            }
        }
        RoomCreateScreen(
            onBack = navigator::back,
            onOpenApplications = { navigator.navigate(Routes.ROOM_APPLICATIONS) },
            onSubmitted = { navigator.navigate(Routes.ROOM_APPLICATIONS) },
            onPickLocation = { navigator.navigate(Routes.locationPick()) },
            viewModel = viewModel,
        )
    }
    composable(Routes.ROOM_APPLICATIONS) {
        RoomApplicationsScreen(
            onBack = navigator::back,
            onOpenRoom = { roomId ->
                navigator.navController.navigate(Routes.room(roomId)) {
                    popUpTo(Routes.ROOM_CREATE) { inclusive = false }
                }
            },
        )
    }
}
