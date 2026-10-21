package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.PostSignupItem
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppPrimaryButton

@Composable
internal fun SignupCard(
    item: PostSignupItem,
    needCheckIn: Boolean,
    checkInWindowOpen: Boolean,
    checking: Boolean,
    onManualCheckIn: () -> Unit,
    onEditRemark: () -> Unit,
) {
    val confirmedSignup = item.signupStatus == PostSignupRules.STATUS_CONFIRMED
    val checkedIn = PostSignupRules.hasTimestamp(item.checkedInAt)
    val remark = item.remark.trim()
    val statusText = if (needCheckIn && confirmedSignup) {
        stringResource(PostSignupRules.checkInLabel(checkedIn))
    } else {
        item.signupStatusText.ifBlank {
            when (item.signupStatus) {
                PostSignupRules.STATUS_CONFIRMED -> stringResource(R.string.signup_status_confirmed)
                PostSignupRules.STATUS_WAITLISTED -> stringResource(R.string.signup_status_waitlisted)
                PostSignupRules.STATUS_CANCELLED -> stringResource(R.string.signup_status_cancelled)
                else -> ""
            }
        }
    }
    val (fg, bg) = when {
        item.signupStatus == PostSignupRules.STATUS_WAITLISTED -> AppColors.OrangeStrong to AppColors.OrangeTint
        item.signupStatus == PostSignupRules.STATUS_CANCELLED -> AppColors.TextTertiary to AppColors.NeutralTint
        needCheckIn && confirmedSignup && checkedIn -> AppColors.SuccessStrong to AppColors.SuccessTint
        needCheckIn && confirmedSignup -> AppColors.TextTertiary to AppColors.NeutralTint
        confirmedSignup -> AppColors.SuccessStrong to AppColors.SuccessTint
        else -> AppColors.TextTertiary to AppColors.NeutralTint
    }
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg, AppShapes.Card)
            .padding(horizontal = AppSpacing.L, vertical = 14.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.M),
        ) {
            Text(
                item.userNickname,
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            if (statusText.isNotBlank()) {
                Text(
                    statusText,
                    fontSize = AppType.Caption,
                    color = fg,
                    modifier = Modifier
                        .background(bg, AppShapes.Capsule)
                        .padding(horizontal = AppSpacing.S, vertical = AppSpacing.Xxs),
                )
            }
        }
        Text(
            stringResource(R.string.signup_row_signup_at, item.createdAt),
            fontSize = AppType.Caption,
            color = AppColors.TextTertiary,
            modifier = Modifier.padding(top = AppSpacing.S),
        )
        if (needCheckIn && item.checkedInAt.isNotBlank()) {
            Text(
                stringResource(R.string.signup_row_check_in_at, item.checkedInAt),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = AppSpacing.Xs),
            )
        }
        if (item.cancelledAt.isNotBlank()) {
            Text(
                stringResource(R.string.signup_row_cancelled_at, item.cancelledAt),
                fontSize = AppType.Caption,
                color = AppColors.TextTertiary,
                modifier = Modifier.padding(top = AppSpacing.Xs),
            )
        }
        Text(
            if (remark.isNotEmpty()) {
                stringResource(R.string.signup_row_remark, remark)
            } else {
                stringResource(R.string.signup_add_remark)
            },
            fontSize = AppType.Caption,
            color = AppColors.TextTertiary,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onEditRemark)
                .padding(top = AppSpacing.S),
        )
        if (needCheckIn && checkInWindowOpen && confirmedSignup && !checkedIn) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = AppSpacing.M),
                horizontalArrangement = Arrangement.End,
            ) {
                AppPrimaryButton(
                    text = stringResource(if (checking) R.string.signup_checking else R.string.signup_manual_check_in),
                    onClick = onManualCheckIn,
                    enabled = !checking,
                    fillMaxWidth = false,
                    compact = true,
                )
            }
        }
    }
}
