package com.nearby.chat.feature.navigation

import androidx.compose.runtime.Stable
import androidx.navigation.NavHostController
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.feature.forward.ForwardSession
import com.nearby.chat.feature.main.MainViewModel
import com.nearby.chat.feature.report.ReportTargetType

@Stable
class AppNavigator(
    val navController: NavHostController,
    private val mainViewModel: MainViewModel,
) {
    fun navigate(route: String) = navController.navigate(route)

    fun back() {
        navController.popBackStack()
    }

    fun openForward(session: ForwardSession) {
        if (mainViewModel.startForward(session)) navController.navigate(Routes.FORWARD)
    }

    fun forwardMessage(message: MessageItem) {
        openForward(ForwardSession(contentType = message.contentType, content = message.content))
    }

    fun reportMessage(message: MessageItem, fallbackTitle: String) {
        navController.navigate(
            Routes.report(
                targetType = ReportTargetType.MESSAGE,
                targetId = message.id,
                title = message.fromName.ifBlank { fallbackTitle },
                avatar = message.fromAvatar,
            ),
        )
    }

    fun pickShareLocation() = navController.navigate(Routes.locationPick(Routes.PICK_FOR_SHARE))

    fun logoutToLogin() {
        navController.navigate(Routes.LOGIN) {
            popUpTo(Routes.MAIN) { inclusive = true }
        }
    }
}
