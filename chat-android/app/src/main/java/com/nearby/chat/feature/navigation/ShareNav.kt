package com.nearby.chat.feature.navigation

import androidx.navigation.NavController
import com.nearby.chat.core.share.ShareLinks

fun NavController.navigateShareTarget(target: ShareLinks.Target) {
    when (target) {
        is ShareLinks.Target.Post -> navigate(Routes.postDetail(target.id)) {
            launchSingleTop = true
        }

        is ShareLinks.Target.Room -> navigate(Routes.room(target.id)) {
            launchSingleTop = true
        }

        is ShareLinks.Target.User -> navigate(Routes.userProfile(target.id)) {
            launchSingleTop = true
        }

        is ShareLinks.Target.PrivateChat -> navigate(Routes.privateChat(target.peerId)) {
            launchSingleTop = true
        }

        is ShareLinks.Target.Download -> {
            resolvePendingShareTarget(target)?.let { navigateShareTarget(it) }
        }
    }
}

fun resolvePendingShareTarget(target: ShareLinks.Target): ShareLinks.Target? {
    return when (target) {
        is ShareLinks.Target.Download -> {
            val path = target.continuePath ?: return null
            ShareLinks.parseAppPath(path)
        }

        else -> target
    }
}
