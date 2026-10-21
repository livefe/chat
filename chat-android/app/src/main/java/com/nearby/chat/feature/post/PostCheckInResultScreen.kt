package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppBackIconButton
import com.nearby.chat.core.ui.AppPrimaryButton
import com.nearby.chat.core.ui.AppSecondaryButton
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.appTopAppBarColors

private val ResultAvatarSize = 52.dp
private val TimelineDotSize = 8.dp
private val TimelineLineHeight = 28.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostCheckInResultScreen(
    onBack: () -> Unit,
    onContinueScan: () -> Unit,
    onOpenSignups: (Long) -> Unit,
    viewModel: PostCheckInResultViewModel = hiltViewModel(),
) {
    val payload = viewModel.payload
    val isFail = payload?.kind == CheckInNavResult.KIND_FAIL
    val isAlready = payload?.kind == CheckInNavResult.KIND_ALREADY
    val statusText = stringResource(
        when (payload?.kind) {
            CheckInNavResult.KIND_FAIL -> R.string.check_in_failed
            CheckInNavResult.KIND_ALREADY -> R.string.check_in_already
            else -> R.string.check_in_success
        },
    )
    val signupLabel = stringResource(R.string.signup_action)
    val onsiteLabel = stringResource(R.string.check_in_onsite)
    val operatorText = payload?.operatorNickname?.takeIf { it.isNotBlank() }
        ?.let { stringResource(R.string.check_in_operator, it) }
        .orEmpty()
    val timeline = buildList {
        if (!isFail && payload != null) {
            if (payload.signupAt.isNotBlank()) {
                add(CheckInTimelineItem(signupLabel, payload.signupAt, "", false))
            }
            if (payload.checkedInAtText.isNotBlank()) {
                add(
                    CheckInTimelineItem(
                        label = onsiteLabel,
                        time = payload.checkedInAtText,
                        meta = operatorText,
                        active = true,
                    ),
                )
            }
        }
    }

    Scaffold(
        containerColor = AppColors.PageBg,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        stringResource(R.string.check_in_result_title),
                        fontSize = AppType.PageTitle,
                        fontWeight = FontWeight.SemiBold,
                        color = AppColors.TextPrimary,
                    )
                },
                navigationIcon = { AppBackIconButton(onClick = onBack) },
                colors = appTopAppBarColors(AppColors.PageBg),
            )
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = AppSpacing.M)
                .padding(top = AppSpacing.M, bottom = AppSpacing.L),
        ) {
            when {
                payload == null -> FailCard(
                    stringResource(R.string.check_in_data_missing),
                    stringResource(R.string.check_in_rescan),
                )

                isFail -> FailCard(
                    statusText,
                    payload.message.ifBlank { stringResource(R.string.check_in_fail_hint) },
                )

                else -> {
                    UserSummaryCard(
                        nickname = payload.userNickname,
                        avatarUrl = payload.userAvatarUrl,
                        postTitle = payload.postTitle,
                        statusText = statusText,
                        isAlready = isAlready,
                    )
                    if (timeline.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(10.dp))
                        TimelineCard(timeline)
                    }
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            AppPrimaryButton(
                text = stringResource(R.string.check_in_continue),
                onClick = onContinueScan,
                modifier = Modifier.fillMaxWidth(),
            )
            if (viewModel.postId > 0L) {
                Spacer(modifier = Modifier.height(AppSpacing.M))
                AppSecondaryButton(
                    text = stringResource(R.string.check_in_view_signups),
                    onClick = { onOpenSignups(viewModel.postId) },
                    modifier = Modifier.fillMaxWidth(),
                )
            }
        }
    }
}

@Composable
private fun UserSummaryCard(
    nickname: String,
    avatarUrl: String,
    postTitle: String,
    statusText: String,
    isAlready: Boolean,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg, AppShapes.Card)
            .padding(AppSpacing.L),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        UserAvatar(url = avatarUrl, size = ResultAvatarSize)
        Column(
            modifier = Modifier
                .weight(1f)
                .padding(horizontal = AppSpacing.M),
        ) {
            Text(
                text = nickname.ifBlank { stringResource(R.string.user_default_name) },
                color = AppColors.TextPrimary,
                fontSize = AppType.PageTitle,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (postTitle.isNotBlank()) {
                Text(
                    text = postTitle,
                    color = AppColors.TextTertiary,
                    fontSize = AppType.Meta,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.padding(top = AppSpacing.Xs),
                )
            }
        }
        Text(
            text = statusText,
            color = if (isAlready) AppColors.WarningStrong else AppColors.SuccessStrong,
            fontSize = AppType.Caption,
            fontWeight = FontWeight.Medium,
            modifier = Modifier
                .background(
                    if (isAlready) AppColors.WarningStrong.copy(alpha = 0.12f) else AppColors.SuccessTint,
                    AppShapes.Capsule,
                )
                .padding(horizontal = AppSpacing.S, vertical = AppSpacing.Xxs),
        )
    }
}

@Composable
private fun TimelineCard(items: List<CheckInTimelineItem>) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg, AppShapes.Card)
            .padding(AppSpacing.L),
    ) {
        Text(
            text = stringResource(R.string.check_in_timeline),
            color = AppColors.TextSecondary,
            fontSize = AppType.Secondary,
            fontWeight = FontWeight.Medium,
        )
        Spacer(modifier = Modifier.height(14.dp))
        items.forEachIndexed { index, item ->
            TimelineRow(item = item, showLine = index < items.lastIndex)
            if (index < items.lastIndex) Spacer(modifier = Modifier.height(AppSpacing.L))
        }
    }
}

@Composable
private fun TimelineRow(item: CheckInTimelineItem, showLine: Boolean) {
    Row(modifier = Modifier.fillMaxWidth()) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.width(AppSpacing.M),
        ) {
            Box(
                modifier = Modifier
                    .size(TimelineDotSize)
                    .background(
                        if (item.active) AppColors.Brand else AppColors.TimelineDot,
                        CircleShape,
                    ),
            )
            if (showLine) {
                Box(
                    modifier = Modifier
                        .width(AppSpacing.Xxs)
                        .height(TimelineLineHeight)
                        .padding(top = AppSpacing.Xs)
                        .background(AppColors.TimelineLine),
                )
            }
        }
        Column(modifier = Modifier.padding(start = AppSpacing.M)) {
            Text(item.label, color = AppColors.TextPrimary, fontSize = AppType.Subhead, fontWeight = FontWeight.Medium)
            Text(
                item.time,
                color = AppColors.TextSecondary,
                fontSize = AppType.Meta,
                modifier = Modifier.padding(top = AppSpacing.Xs),
            )
            if (item.meta.isNotBlank()) {
                Text(
                    item.meta,
                    color = AppColors.TextTertiary,
                    fontSize = AppType.Meta,
                    modifier = Modifier.padding(top = AppSpacing.Xs),
                )
            }
        }
    }
}

@Composable
private fun FailCard(title: String, message: String) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg, AppShapes.Card)
            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.Xl),
    ) {
        Text(title, color = AppColors.DangerDeep, fontSize = AppType.PageTitle, fontWeight = FontWeight.SemiBold)
        Text(
            message,
            color = AppColors.TextSecondary,
            fontSize = AppType.Secondary,
            lineHeight = 20.sp,
            modifier = Modifier.padding(top = 10.dp),
        )
    }
}
