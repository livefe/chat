package com.nearby.chat.feature.me

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.AppSecondaryButton
import com.nearby.chat.core.ui.UserAvatar

private val MeAvatarSize = 56.dp

@Composable
fun MeScreen(
    onEditProfile: () -> Unit,
    onMyPublished: () -> Unit,
    onMyJoined: () -> Unit,
    onSupport: () -> Unit,
    onFeedback: () -> Unit,
    onSettings: () -> Unit,
    onLogout: () -> Unit,
    viewModel: MeViewModel = hiltViewModel(),
) {
    val user by viewModel.user.collectAsStateWithLifecycle()
    val supportUnread by viewModel.supportUnread.collectAsStateWithLifecycle()
    val latestOnLogout by rememberUpdatedState(onLogout)
    var logoutConfirm by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) { viewModel.refresh() }
    LaunchedEffect(viewModel) {
        viewModel.loggedOut.collect { latestOnLogout() }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppColors.PageBg)
            .statusBarsPadding()
            .verticalScroll(rememberScrollState()),
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier
                .padding(start = AppSpacing.L, end = AppSpacing.L, top = AppSpacing.M)
                .fillMaxWidth()
                .clip(AppShapes.Card)
                .background(AppColors.CardBg)
                .clickable(onClick = onEditProfile)
                .padding(horizontal = AppSpacing.L, vertical = 18.dp),
        ) {
            UserAvatar(
                url = user?.avatarUrl,
                size = MeAvatarSize,
            )
            Column(
                modifier = Modifier
                    .weight(1f)
                    .padding(start = 14.dp),
            ) {
                Text(
                    user?.nickname ?: stringResource(R.string.me_not_logged_in),
                    fontSize = AppType.Profile,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                )
                val mobile = user?.mobile.orEmpty()
                if (mobile.isNotBlank()) {
                    Text(
                        mobile,
                        fontSize = AppType.Meta,
                        lineHeight = AppType.LineMeta,
                        color = AppColors.TextSecondary,
                        modifier = Modifier.padding(top = 6.dp),
                    )
                }
            }
            Icon(
                Icons.AutoMirrored.Outlined.KeyboardArrowRight,
                contentDescription = null,
                tint = AppColors.TextSecondary,
                modifier = Modifier.size(AppIconSize.Cell),
            )
        }

        Spacer(Modifier.height(AppSpacing.M))

        MeSection {
            MeCell(stringResource(R.string.my_posts_published), onClick = onMyPublished)
            MeDivider()
            MeCell(stringResource(R.string.my_posts_joined), onClick = onMyJoined)
        }

        Spacer(Modifier.height(AppSpacing.M))

        MeSection {
            MeCell(stringResource(R.string.me_feedback), onClick = onFeedback)
            MeDivider()
            MeCell(stringResource(R.string.me_support), onClick = onSupport, badge = supportUnread)
        }

        Spacer(Modifier.height(AppSpacing.M))

        MeSection {
            MeCell(stringResource(R.string.me_settings), onClick = onSettings)
        }

        AppSecondaryButton(
            text = stringResource(R.string.me_logout),
            onClick = { logoutConfirm = true },
            modifier = Modifier.padding(horizontal = AppSpacing.L, vertical = AppSpacing.Xxl),
        )
    }

    if (logoutConfirm) {
        AppConfirmDialog(
            title = stringResource(R.string.me_logout),
            message = stringResource(R.string.me_logout_confirm),
            onConfirm = {
                logoutConfirm = false
                viewModel.logout()
            },
            onDismiss = { logoutConfirm = false },
        )
    }
}

@Composable
private fun MeDivider() {
    HorizontalDivider(color = AppColors.Divider, modifier = Modifier.padding(start = AppSpacing.L))
}

@Composable
private fun MeSection(content: @Composable () -> Unit) {
    Column(
        modifier = Modifier
            .padding(horizontal = AppSpacing.L)
            .fillMaxWidth()
            .clip(AppShapes.Card)
            .background(AppColors.CardBg),
        content = { content() },
    )
}
