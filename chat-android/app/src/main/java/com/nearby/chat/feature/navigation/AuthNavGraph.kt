package com.nearby.chat.feature.navigation

import androidx.compose.ui.res.stringResource
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.nearby.chat.R
import com.nearby.chat.core.h5.H5Urls
import com.nearby.chat.feature.auth.LoginScreen
import com.nearby.chat.feature.h5.H5WebScreen

internal fun NavGraphBuilder.authGraph(
    navigator: AppNavigator,
    onLoggedIn: () -> Unit,
) {
    composable(Routes.LOGIN) {
        val userAgreementTitle = stringResource(R.string.h5_user_agreement)
        val privacyPolicyTitle = stringResource(R.string.h5_privacy_policy)
        LoginScreen(
            onLoggedIn = onLoggedIn,
            onOpenUserAgreement = {
                navigator.navigate(Routes.h5Web(Routes.H5_AGREEMENT_USER, userAgreementTitle))
            },
            onOpenPrivacyPolicy = {
                navigator.navigate(Routes.h5Web(Routes.H5_AGREEMENT_PRIVACY, privacyPolicyTitle))
            },
        )
    }
    composable(
        route = Routes.H5_WEB,
        arguments = listOf(
            navArgument("path") { type = NavType.StringType },
            navArgument("title") {
                type = NavType.StringType
                defaultValue = ""
            },
        ),
    ) { entry ->
        val path = entry.arguments?.getString("path").orEmpty()
        H5WebScreen(
            url = H5Urls.page(path),
            title = entry.arguments?.getString("title").orEmpty(),
            onBack = navigator::back,
        )
    }
}
