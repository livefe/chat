package com.nearby.chat.feature.post

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.RadioButton
import androidx.compose.material3.RadioButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.format.formatUnixTime
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.feature.post.ui.PostClickRow
import com.nearby.chat.feature.post.ui.PostFormField
import com.nearby.chat.feature.post.ui.PostSwitchRow

internal val PostFormShortLabelWidth = 72.dp

@Composable
internal fun PostEventSections(
    state: PostCreateUiState,
    viewModel: PostCreateViewModel,
    onPickDateTime: (PostDateTimeTarget) -> Unit,
) {
    val context = LocalContext.current
    val pleaseSelect = stringResource(R.string.common_please_select)
    val isCharge = state.feeKind == FeeKind.Charge
    PostFormCard {
        PostSwitchRow(stringResource(R.string.post_schedule), state.enableSchedule, viewModel::onEnableSchedule)
        if (state.enableSchedule) {
            PostFormDivider()
            PostClickRow(
                label = stringResource(R.string.post_start_time),
                value = formatUnixTime(context, state.startAt),
                placeholder = pleaseSelect,
                onClick = { onPickDateTime(PostDateTimeTarget.Start) },
            )
            PostFormDivider()
            PostClickRow(
                label = stringResource(R.string.post_end_time),
                value = formatUnixTime(context, state.endAt),
                placeholder = pleaseSelect,
                onClick = { onPickDateTime(PostDateTimeTarget.End) },
            )
        }
    }
    PostFormCard {
        PostSwitchRow(stringResource(R.string.post_fee_setting), state.enableFee, viewModel::onEnableFee)
        if (state.enableFee) {
            PostFormDivider()
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = AppSpacing.L, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    stringResource(R.string.post_fee_kind),
                    fontSize = AppType.Secondary,
                    color = AppColors.TextPrimary,
                    modifier = Modifier.width(PostFormShortLabelWidth),
                )
                FeeKindOption(stringResource(R.string.fee_kind_charge), isCharge) {
                    viewModel.onFeeKind(FeeKind.Charge)
                }
                FeeKindOption(stringResource(R.string.fee_kind_reward), !isCharge) {
                    viewModel.onFeeKind(FeeKind.Reward)
                }
            }
            PostFormDivider()
            PostFormField(
                label = stringResource(if (isCharge) R.string.post_fee_amount_charge else R.string.post_fee_amount_reward),
                value = state.feeAmount,
                onValueChange = viewModel::onFeeAmount,
                placeholder = stringResource(R.string.post_amount_hint),
                singleLine = true,
                keyboardType = KeyboardType.Decimal,
                suffix = stringResource(R.string.unit_yuan),
            )
            PostFormDivider()
            PostFormField(
                label = stringResource(if (isCharge) R.string.post_fee_desc_charge else R.string.post_fee_desc_reward),
                value = state.feeDescription,
                onValueChange = viewModel::onFeeDescription,
                placeholder = stringResource(R.string.common_optional),
                singleLine = true,
            )
        }
    }
    PostFormCard {
        PostSwitchRow(stringResource(R.string.post_need_signup), state.needSignup, viewModel::onNeedSignup)
        if (state.needSignup) {
            PostFormDivider()
            PostClickRow(
                label = stringResource(R.string.post_signup_deadline),
                value = formatUnixTime(context, state.signupDeadlineAt),
                placeholder = pleaseSelect,
                onClick = { onPickDateTime(PostDateTimeTarget.SignupDeadline) },
            )
            PostFormDivider()
            PostSwitchRow(
                stringResource(R.string.post_limit_participants),
                state.limitParticipants,
                viewModel::onLimitParticipants,
            )
            if (state.limitParticipants) {
                PostFormDivider()
                PostFormField(
                    label = stringResource(R.string.post_max_participants),
                    value = state.maxParticipants,
                    onValueChange = viewModel::onMaxParticipants,
                    placeholder = stringResource(R.string.post_max_hint),
                    singleLine = true,
                    keyboardType = KeyboardType.Number,
                    suffix = stringResource(R.string.unit_person),
                )
                PostFormDivider()
                PostSwitchRow(stringResource(R.string.post_allow_waitlist), state.allowWaitlist, viewModel::onAllowWaitlist)
            }
            PostFormDivider()
            PostSwitchRow(stringResource(R.string.post_need_check_in), state.needCheckIn, viewModel::onNeedCheckIn)
        }
    }
}

@Composable
private fun FeeKindOption(label: String, selected: Boolean, onClick: () -> Unit) {
    val focusManager = LocalFocusManager.current
    val keyboardController = LocalSoftwareKeyboardController.current
    val handleClick = {
        focusManager.clearFocus(force = true)
        keyboardController?.hide()
        onClick()
    }
    Row(
        modifier = Modifier
            .clickable(onClick = handleClick)
            .padding(end = AppSpacing.L),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(AppSpacing.Xs),
    ) {
        RadioButton(
            selected = selected,
            onClick = handleClick,
            colors = RadioButtonDefaults.colors(selectedColor = AppColors.Brand),
        )
        Text(label, fontSize = AppType.Secondary, color = AppColors.TextPrimary)
    }
}
