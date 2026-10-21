package com.nearby.chat.feature.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.MoreHoriz
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.model.PublicUserProfile
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.UserAvatar

private val ProfileAvatarSize = 72.dp
private val ProfileNameSize = 20.sp
private val IconButtonSlot = 48.dp
private val LoadingPadding = 32.dp
private val StatusTopPadding = 6.dp
private val MessageButtonMinWidth = 96.dp

@Composable
internal fun UserProfileHeader(
    state: UserProfileUiState,
    canReport: Boolean,
    canMessage: Boolean,
    onBack: () -> Unit,
    onMenu: () -> Unit,
    onChat: (PublicUserProfile) -> Unit,
) {
    val profile = state.profile
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg)
            .statusBarsPadding()
            .padding(bottom = AppSpacing.Xl),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = AppSpacing.Xs, vertical = AppSpacing.Xxs),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            AppBackIconButton(onClick = onBack)
            if (canReport) {
                IconButton(onClick = onMenu) {
                    Icon(
                        Icons.Outlined.MoreHoriz,
                        contentDescription = stringResource(R.string.common_more),
                        tint = AppColors.TextPrimary,
                        modifier = Modifier.size(AppIconSize.Cell),
                    )
                }
            } else {
                Spacer(Modifier.size(IconButtonSlot))
            }
        }

        when {
            state.profileLoading && profile == null -> {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = LoadingPadding),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator(color = AppColors.Brand, strokeWidth = 2.dp)
                }
            }

            profile != null -> ProfileSummary(profile, canMessage, onChat)

            state.loadFailed -> {
                Text(
                    stringResource(R.string.common_load_failed),
                    color = AppColors.TextTertiary,
                    modifier = Modifier
                        .align(Alignment.CenterHorizontally)
                        .padding(AppSpacing.Xxl),
                )
            }
        }
    }
}

@Composable
private fun ProfileSummary(
    profile: PublicUserProfile,
    canMessage: Boolean,
    onChat: (PublicUserProfile) -> Unit,
) {
    val joinLabel = joinDateText(profile.createdAt)
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        UserAvatar(
            url = profile.avatarUrl,
            size = ProfileAvatarSize,
        )
        Text(
            profile.nickname,
            fontSize = ProfileNameSize,
            fontWeight = FontWeight.SemiBold,
            color = AppColors.TextPrimary,
            modifier = Modifier.padding(top = AppSpacing.M),
        )
        if (joinLabel.isNotBlank()) {
            Text(
                stringResource(R.string.profile_joined_at, joinLabel),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = StatusTopPadding),
            )
        }
        if (profile.deactivated) {
            Text(
                stringResource(R.string.profile_deactivated),
                fontSize = AppType.Caption,
                color = AppColors.Danger,
                modifier = Modifier.padding(top = StatusTopPadding),
            )
        } else if (profile.blockedByMe) {
            Text(
                stringResource(R.string.profile_blocked_by_me),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = StatusTopPadding),
            )
        }
        if (canMessage) {
            AppPrimaryButton(
                text = stringResource(R.string.profile_send_message),
                onClick = { onChat(profile) },
                compact = true,
                modifier = Modifier
                    .padding(top = AppSpacing.M)
                    .widthIn(min = MessageButtonMinWidth),
            )
        }
    }
}
