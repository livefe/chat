package com.nearby.chat.feature.post

import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import com.nearby.chat.R
import com.nearby.chat.feature.post.ui.PostFormField
import com.nearby.chat.feature.post.ui.PostSwitchRow

@Composable
internal fun PostHelpSections(state: PostCreateUiState, viewModel: PostCreateViewModel) {
    PostFormCard {
        PostSwitchRow(stringResource(R.string.post_help_reward), state.helpReward, viewModel::onHelpReward)
        if (state.helpReward) {
            PostFormDivider()
            PostFormField(
                label = stringResource(R.string.post_fee_amount_reward),
                value = state.helpRewardAmount,
                onValueChange = viewModel::onHelpRewardAmount,
                placeholder = stringResource(R.string.post_amount_hint),
                singleLine = true,
                keyboardType = KeyboardType.Decimal,
                suffix = stringResource(R.string.unit_yuan),
            )
            PostFormDivider()
            PostFormField(
                label = stringResource(R.string.post_fee_desc_reward),
                value = state.helpRewardDesc,
                onValueChange = viewModel::onHelpRewardDesc,
                placeholder = stringResource(R.string.common_optional),
                singleLine = true,
            )
        }
    }
}
