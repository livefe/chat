package com.nearby.chat.feature.post

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Reply
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppSecondaryButton

@Composable
internal fun PostDetailBottomBar(
    post: PostItem,
    onShare: () -> Unit,
    onSignup: () -> Unit,
    onCancelSignup: () -> Unit,
    onOpenChat: () -> Unit,
    showDiscussion: Boolean,
) {
    val needSignup = post.extras?.needSignup == true
    val signedUp = post.mySignupStatus == PostSignupRules.STATUS_CONFIRMED ||
        post.mySignupStatus == PostSignupRules.STATUS_WAITLISTED
    val deadline = post.extras?.signupDeadlineAt ?: 0L
    val signupClosed = needSignup && PostSignupRules.isSignupDeadlinePassed(deadline)
    val eventEnded = needSignup && PostSignupRules.isEventEnded(post.endAt)
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .navigationBarsPadding()
            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(
            modifier = Modifier
                .clickable(onClick = onShare)
                .padding(vertical = 10.dp, horizontal = AppSpacing.Xs),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Icon(
                Icons.AutoMirrored.Outlined.Reply,
                contentDescription = stringResource(R.string.common_share),
                tint = AppColors.TextPrimary,
                modifier = Modifier
                    .size(AppIconSize.Inline)
                    .graphicsLayer { scaleX = -1f },
            )
            Text(
                stringResource(R.string.common_share),
                fontSize = AppType.Secondary,
                color = AppColors.TextPrimary,
                lineHeight = AppType.Secondary,
            )
        }
        Row(
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (needSignup) {
                when {
                    signedUp -> AppSecondaryButton(
                        text = stringResource(
                            if (post.mySignupStatus == PostSignupRules.STATUS_WAITLISTED) {
                                R.string.signup_cancel_waitlist_action
                            } else {
                                R.string.signup_cancel_action
                            },
                        ),
                        onClick = onCancelSignup,
                        fillMaxWidth = false,
                        compact = true,
                    )

                    signupClosed || eventEnded -> AppSecondaryButton(
                        text = stringResource(
                            if (signupClosed) R.string.signup_deadline_passed else R.string.phase_ended,
                        ),
                        onClick = {},
                        enabled = false,
                        fillMaxWidth = false,
                        compact = true,
                    )

                    else -> AppPrimaryButton(
                        text = stringResource(R.string.signup_action),
                        onClick = onSignup,
                        fillMaxWidth = false,
                        compact = true,
                    )
                }
                if (showDiscussion) {
                    AppSecondaryButton(
                        text = stringResource(R.string.post_discussion),
                        onClick = onOpenChat,
                        fillMaxWidth = false,
                        compact = true,
                    )
                }
            } else if (showDiscussion) {
                AppPrimaryButton(
                    text = stringResource(R.string.post_discussion),
                    onClick = onOpenChat,
                    fillMaxWidth = false,
                    compact = true,
                )
            }
        }
    }
}
