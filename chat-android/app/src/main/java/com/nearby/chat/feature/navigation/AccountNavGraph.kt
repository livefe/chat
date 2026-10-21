package com.nearby.chat.feature.navigation

import androidx.compose.runtime.getValue
import androidx.compose.ui.res.stringResource
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.nearby.chat.R
import com.nearby.chat.feature.account.AccountDeactivateScreen
import com.nearby.chat.feature.account.BlockListScreen
import com.nearby.chat.feature.account.EditMobileScreen
import com.nearby.chat.feature.account.EditNicknameScreen
import com.nearby.chat.feature.account.EditProfileScreen
import com.nearby.chat.feature.account.SettingsScreen
import com.nearby.chat.feature.feedback.FeedbackDetailScreen
import com.nearby.chat.feature.feedback.FeedbackDetailViewModel
import com.nearby.chat.feature.feedback.FeedbackListScreen
import com.nearby.chat.feature.feedback.FeedbackScreen
import com.nearby.chat.feature.profile.UserProfileScreen
import com.nearby.chat.feature.report.ReportScreen
import com.nearby.chat.feature.report.ReportTargetType
import com.nearby.chat.feature.update.AppUpdateViewModel

internal fun NavGraphBuilder.accountGraph(
    navigator: AppNavigator,
    updateViewModel: AppUpdateViewModel,
) {
    composable(
        route = Routes.USER_PROFILE,
        arguments = listOf(navArgument("userId") { type = NavType.LongType }),
    ) { entry ->
        UserProfileScreen(
            userId = entry.arguments?.getLong("userId") ?: 0L,
            onBack = navigator::back,
            onChat = { id, name -> navigator.navigate(Routes.privateChat(id, name)) },
            onReport = { id, nickname, avatar ->
                navigator.navigate(Routes.report(ReportTargetType.USER, id, nickname, avatar))
            },
            onOpenPost = { navigator.navigate(Routes.postDetail(it)) },
        )
    }
    composable(Routes.EDIT_PROFILE) {
        EditProfileScreen(
            onBack = navigator::back,
            onEditNickname = { navigator.navigate(Routes.EDIT_NICKNAME) },
            onEditMobile = { navigator.navigate(Routes.EDIT_MOBILE) },
        )
    }
    composable(Routes.EDIT_NICKNAME) {
        EditNicknameScreen(onBack = navigator::back)
    }
    composable(Routes.EDIT_MOBILE) {
        EditMobileScreen(onBack = navigator::back)
    }
    composable(Routes.FEEDBACK) {
        FeedbackScreen(
            onBack = navigator::back,
            onOpenRecords = { navigator.navigate(Routes.FEEDBACK_RECORDS) },
        )
    }
    composable(Routes.FEEDBACK_RECORDS) {
        FeedbackListScreen(
            onBack = navigator::back,
            onOpenDetail = { navigator.navigate(Routes.feedbackDetail(it)) },
        )
    }
    composable(
        route = Routes.FEEDBACK_DETAIL,
        arguments = listOf(navArgument(FeedbackDetailViewModel.ARG_ID) { type = NavType.LongType }),
    ) {
        FeedbackDetailScreen(onBack = navigator::back)
    }
    composable(Routes.SETTINGS) {
        val updateState by updateViewModel.uiState.collectAsStateWithLifecycle()
        val userAgreementTitle = stringResource(R.string.h5_user_agreement)
        val privacyPolicyTitle = stringResource(R.string.h5_privacy_policy)
        SettingsScreen(
            onBack = navigator::back,
            onOpenBlocks = { navigator.navigate(Routes.BLOCK_LIST) },
            onOpenDeactivate = { navigator.navigate(Routes.ACCOUNT_DEACTIVATE) },
            onOpenUserAgreement = {
                navigator.navigate(Routes.h5Web(Routes.H5_AGREEMENT_USER, userAgreementTitle))
            },
            onOpenPrivacyPolicy = {
                navigator.navigate(Routes.h5Web(Routes.H5_AGREEMENT_PRIVACY, privacyPolicyTitle))
            },
            onCheckUpdate = updateViewModel::checkManual,
            checkingUpdate = updateState.checking,
        )
    }
    composable(Routes.BLOCK_LIST) {
        BlockListScreen(
            onBack = navigator::back,
            onOpenProfile = { userId, _, _ -> navigator.navigate(Routes.userProfile(userId)) },
        )
    }
    composable(Routes.ACCOUNT_DEACTIVATE) {
        AccountDeactivateScreen(
            onBack = navigator::back,
            onDone = navigator::logoutToLogin,
        )
    }
    composable(
        route = Routes.REPORT,
        arguments = listOf(
            navArgument("targetType") {
                type = NavType.IntType
                defaultValue = ReportTargetType.USER
            },
            navArgument("targetId") {
                type = NavType.LongType
                defaultValue = 0L
            },
            navArgument("title") {
                type = NavType.StringType
                defaultValue = ""
                nullable = true
            },
            navArgument("avatar") {
                type = NavType.StringType
                defaultValue = ""
                nullable = true
            },
        ),
    ) { entry ->
        ReportScreen(
            targetType = entry.arguments?.getInt("targetType") ?: ReportTargetType.USER,
            targetId = entry.arguments?.getLong("targetId") ?: 0L,
            previewTitle = entry.arguments?.getString("title").orEmpty(),
            previewAvatar = entry.arguments?.getString("avatar").orEmpty(),
            onBack = navigator::back,
        )
    }
}
