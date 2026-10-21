package com.nearby.chat.feature.post

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

@Composable
internal fun PostDetailExtras(post: PostItem) {
    val context = LocalContext.current
    val extras = post.extras
    val idlePrice = PostDetailFormatter.idlePrice(post)?.asString()
    val hasSchedule = post.startAt > 0 || post.endAt > 0
    val needSignup = extras?.needSignup == true

    Column {
        if (idlePrice != null) {
            ExtraRow(label = stringResource(R.string.extra_price), value = idlePrice)
        }
        if (hasSchedule) {
            if (idlePrice != null) ExtraSectionDivider()
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                if (post.startAt > 0) {
                    ExtraRow(
                        label = stringResource(if (post.endAt > 0) R.string.extra_start else R.string.extra_time),
                        value = PostDetailFormatter.unixTime(context, post.startAt),
                    )
                }
                if (post.endAt > 0) {
                    ExtraRow(
                        label = stringResource(if (post.startAt > 0) R.string.extra_end else R.string.extra_time),
                        value = PostDetailFormatter.unixTime(context, post.endAt),
                    )
                }
            }
        }
        if (needSignup && extras != null) {
            if (idlePrice != null || hasSchedule) ExtraSectionDivider()
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                val max = extras.maxParticipants
                if (max > 0) {
                    ExtraRow(
                        label = stringResource(R.string.extra_quota),
                        value = PostDetailFormatter.quota(extras).asString(),
                    )
                }
                if (extras.signupDeadlineAt > 0) {
                    ExtraRow(
                        label = stringResource(R.string.extra_deadline),
                        value = PostDetailFormatter.unixTime(context, extras.signupDeadlineAt),
                    )
                }
                ExtraRow(
                    label = stringResource(R.string.extra_signup),
                    value = PostDetailFormatter.signupSummary(post, max).asString(),
                )
            }
        }
        if (extras != null && PostDetailFormatter.hasFee(extras)) {
            if (idlePrice != null || hasSchedule || needSignup) ExtraSectionDivider()
            Column(verticalArrangement = Arrangement.spacedBy(AppSpacing.S)) {
                ExtraRow(
                    label = stringResource(R.string.extra_amount),
                    value = PostDetailFormatter.fee(extras).asString(),
                    valueColor = if (extras.feeType == PostDetailFormatter.FEE_TYPE_REWARD) {
                        AppColors.Reward
                    } else {
                        AppColors.Warning
                    },
                    valueWeight = FontWeight.SemiBold,
                )
                if (extras.feeDescription.isNotBlank()) {
                    Text(
                        extras.feeDescription,
                        fontSize = AppType.Meta,
                        color = AppColors.TextSecondary,
                        lineHeight = 20.sp,
                        modifier = Modifier.padding(start = 52.dp),
                    )
                }
            }
        }
    }
}

@Composable
private fun ExtraSectionDivider() {
    HorizontalDivider(
        color = AppColors.BorderSoft,
        modifier = Modifier.padding(vertical = AppSpacing.M),
    )
}

@Composable
private fun ExtraRow(
    label: String,
    value: String,
    valueColor: Color = AppColors.TextPrimary,
    valueWeight: FontWeight = FontWeight.Normal,
) {
    Row(
        verticalAlignment = Alignment.Top,
        horizontalArrangement = Arrangement.spacedBy(AppSpacing.M),
    ) {
        Text(
            label,
            fontSize = AppType.Meta,
            color = AppColors.TextTertiary,
            modifier = Modifier
                .width(40.dp)
                .padding(top = 1.dp),
        )
        Text(
            value,
            fontSize = AppType.Subhead,
            color = valueColor,
            fontWeight = valueWeight,
            lineHeight = 22.sp,
            modifier = Modifier.weight(1f),
        )
    }
}
