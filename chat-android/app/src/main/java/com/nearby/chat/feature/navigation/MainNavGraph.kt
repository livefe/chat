package com.nearby.chat.feature.navigation

import androidx.compose.runtime.getValue
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraphBuilder
import androidx.navigation.compose.composable
import com.nearby.chat.R
import com.nearby.chat.feature.chat.NearbyChatViewModel
import com.nearby.chat.feature.main.MainScaffold

internal fun NavGraphBuilder.mainGraph(navigator: AppNavigator) {
    composable(Routes.MAIN) { entry ->
        val handle = entry.savedStateHandle
        val pickLat by handle.getStateFlow<Double?>(NavResultKeys.PICK_LAT, null).collectAsStateWithLifecycle()
        val pickLng by handle.getStateFlow<Double?>(NavResultKeys.PICK_LNG, null).collectAsStateWithLifecycle()
        val pickPlace by handle.getStateFlow<String?>(NavResultKeys.PICK_PLACE, null).collectAsStateWithLifecycle()
        val messageLabel = stringResource(R.string.report_target_message)
        val chatViewModel: NearbyChatViewModel = hiltViewModel()
        ConsumeShareLocation(entry) { chatViewModel.thread.sendLocation(it.lat, it.lng, it.placeName) }
        MainScaffold(
            onOpenPrivateChat = { peerId, name -> navigator.navigate(Routes.privateChat(peerId, name)) },
            onOpenSupport = { navigator.navigate(Routes.SUPPORT_CHAT) },
            onOpenPostDetail = { navigator.navigate(Routes.postDetail(it)) },
            onOpenRoom = { navigator.navigate(Routes.room(it)) },
            onOpenNearbyRooms = { navigator.navigate(Routes.NEARBY_ROOMS) },
            onOpenEditProfile = { navigator.navigate(Routes.EDIT_PROFILE) },
            onOpenMyPublished = { navigator.navigate(Routes.myPosts(Routes.MY_POSTS_PUBLISHED)) },
            onOpenMyJoined = { navigator.navigate(Routes.myPosts(Routes.MY_POSTS_JOINED)) },
            onOpenFeedback = { navigator.navigate(Routes.FEEDBACK) },
            onOpenSettings = { navigator.navigate(Routes.SETTINGS) },
            onOpenUserProfile = { navigator.navigate(Routes.userProfile(it)) },
            onOpenPostCreate = { lat, lng, place ->
                val navController = navigator.navController
                navController.navigate(Routes.postCreate())
                if (lat != null && lng != null) {
                    navController.currentBackStackEntry?.savedStateHandle?.apply {
                        set(NavResultKeys.PICK_LAT, lat)
                        set(NavResultKeys.PICK_LNG, lng)
                        set(NavResultKeys.PICK_PLACE, place.orEmpty())
                    }
                }
            },
            onOpenLocationPick = { navigator.navigate(Routes.locationPick()) },
            onLogout = navigator::logoutToLogin,
            chatViewModel = chatViewModel,
            onOpenShareLocation = navigator::pickShareLocation,
            onReportMessage = { navigator.reportMessage(it, messageLabel) },
            onOpenForward = navigator::forwardMessage,
            locationPickLat = pickLat,
            locationPickLng = pickLng,
            locationPickPlace = pickPlace,
            onLocationPickConsumed = {
                handle[NavResultKeys.PICK_LAT] = null
                handle[NavResultKeys.PICK_LNG] = null
                handle[NavResultKeys.PICK_PLACE] = null
            },
        )
    }
}
