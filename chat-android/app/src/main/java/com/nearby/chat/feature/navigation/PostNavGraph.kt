package com.nearby.chat.feature.navigation

import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.encodeChatPost
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.feature.forward.ForwardSession
import com.nearby.chat.feature.post.MyPostsScreen
import com.nearby.chat.feature.post.PostCheckInCodeScreen
import com.nearby.chat.feature.post.PostCheckInResultScreen
import com.nearby.chat.feature.post.PostCheckInScanScreen
import com.nearby.chat.feature.post.PostCreateScreen
import com.nearby.chat.feature.post.PostCreateViewModel
import com.nearby.chat.feature.post.PostDetailScreen
import com.nearby.chat.feature.post.PostDetailViewModel
import com.nearby.chat.feature.post.PostSignupListScreen
import com.nearby.chat.feature.report.ReportTargetType

internal fun NavGraphBuilder.postGraph(navigator: AppNavigator) {
    composable(
        route = Routes.POST_DETAIL,
        arguments = listOf(navArgument("postId") { type = NavType.LongType }),
    ) { entry ->
        val postId = entry.arguments?.getLong("postId") ?: 0L
        val messageLabel = stringResource(R.string.report_target_message)
        val viewModel: PostDetailViewModel = hiltViewModel()
        ConsumeShareLocation(entry) { viewModel.thread.sendLocation(it.lat, it.lng, it.placeName) }
        PostDetailScreen(
            postId = postId,
            onBack = navigator::back,
            onOpenPublisher = { navigator.navigate(Routes.userProfile(it)) },
            onOpenPost = { id -> if (id != postId) navigator.navigate(Routes.postDetail(id)) },
            onReport = { id, title -> navigator.navigate(Routes.report(ReportTargetType.CONTENT, id, title)) },
            onReportMessage = { navigator.reportMessage(it, messageLabel) },
            onShare = { title, cover, body ->
                navigator.openForward(
                    ForwardSession(
                        contentType = MessageContentTypes.POST,
                        content = encodeChatPost(postId, title, cover),
                        pageTitle = UiText.of(R.string.forward_share_post_title),
                        postId = postId,
                        postTitle = title,
                        cover = cover,
                        shareContent = body,
                        showExternalShare = true,
                    ),
                )
            },
            onForwardMessage = navigator::forwardMessage,
            onPickLocation = navigator::pickShareLocation,
            viewModel = viewModel,
        )
    }
    composable(
        route = Routes.MY_POSTS,
        arguments = listOf(navArgument("mode") { type = NavType.StringType }),
    ) { entry ->
        val handle = entry.savedStateHandle
        val refreshMyPosts by handle.getStateFlow(NavResultKeys.REFRESH_MY_POSTS, false).collectAsStateWithLifecycle()
        MyPostsScreen(
            refreshTick = refreshMyPosts,
            onRefreshConsumed = { handle[NavResultKeys.REFRESH_MY_POSTS] = false },
            onBack = navigator::back,
            onOpenPost = { navigator.navigate(Routes.postDetail(it)) },
            onEditPost = { navigator.navigate(Routes.postCreate(postId = it)) },
            onOpenSignupList = { navigator.navigate(Routes.postSignups(it)) },
            onOpenCheckInCode = { navigator.navigate(Routes.postCheckInCode(it)) },
            onOpenScan = { navigator.navigate(Routes.checkInScan()) },
        )
    }
    composable(
        route = Routes.POST_SIGNUPS,
        arguments = listOf(navArgument("postId") { type = NavType.LongType }),
    ) {
        PostSignupListScreen(onBack = navigator::back)
    }
    composable(
        route = Routes.POST_CHECK_IN_CODE,
        arguments = listOf(navArgument("postId") { type = NavType.LongType }),
    ) {
        PostCheckInCodeScreen(onBack = navigator::back)
    }
    composable(Routes.CHECK_IN_SCAN) {
        PostCheckInScanScreen(
            onBack = navigator::back,
            onCheckedIn = { navigator.navigate(Routes.checkInResult(payload = it.payloadJson)) },
        )
    }
    composable(
        route = Routes.CHECK_IN_RESULT,
        arguments = listOf(
            navArgument("payload") {
                type = NavType.StringType
                defaultValue = ""
            },
        ),
    ) {
        val navController = navigator.navController
        PostCheckInResultScreen(
            onBack = navigator::back,
            onContinueScan = {
                if (!navController.popBackStack(Routes.CHECK_IN_SCAN, inclusive = false)) {
                    navController.navigate(Routes.checkInScan()) {
                        popUpTo(Routes.CHECK_IN_RESULT) { inclusive = true }
                    }
                }
            },
            onOpenSignups = { id ->
                navController.navigate(Routes.postSignups(id)) {
                    popUpTo(Routes.CHECK_IN_SCAN) { inclusive = true }
                }
            },
        )
    }
    composable(
        route = Routes.POST_CREATE,
        arguments = listOf(
            navArgument("roomId") {
                type = NavType.LongType
                defaultValue = 0L
            },
            navArgument("postId") {
                type = NavType.LongType
                defaultValue = 0L
            },
        ),
    ) { entry ->
        val viewModel: PostCreateViewModel = hiltViewModel()
        val roomId = entry.arguments?.getLong("roomId") ?: 0L
        val postId = entry.arguments?.getLong("postId") ?: 0L
        val handle = entry.savedStateHandle
        val pickLat by handle.getStateFlow<Double?>(NavResultKeys.PICK_LAT, null).collectAsStateWithLifecycle()
        val pickLng by handle.getStateFlow<Double?>(NavResultKeys.PICK_LNG, null).collectAsStateWithLifecycle()
        val pickPlace by handle.getStateFlow<String?>(NavResultKeys.PICK_PLACE, null).collectAsStateWithLifecycle()
        LaunchedEffect(pickLat, pickLng, pickPlace) {
            val lat = pickLat
            val lng = pickLng
            if (lat != null && lng != null) {
                viewModel.applyLocation(lat, lng, pickPlace.orEmpty())
                handle[NavResultKeys.PICK_LAT] = null
                handle[NavResultKeys.PICK_LNG] = null
                handle[NavResultKeys.PICK_PLACE] = null
            }
        }
        val navController = navigator.navController
        PostCreateScreen(
            onBack = navigator::back,
            onCreated = { id ->
                if (postId > 0) {
                    navController.previousBackStackEntry?.savedStateHandle?.set(NavResultKeys.REFRESH_MY_POSTS, true)
                    navController.popBackStack()
                } else {
                    navController.popBackStack()
                    navController.navigate(Routes.postDetail(id))
                }
            },
            onPickLocation = { navigator.navigate(Routes.locationPick()) },
            roomId = roomId,
            postId = postId,
            viewModel = viewModel,
        )
    }
}
