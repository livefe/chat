package com.nearby.chat.feature.post

import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import com.nearby.chat.R
import com.nearby.chat.feature.post.ui.PostFormField
import com.nearby.chat.feature.post.ui.PostSwitchRow

@Composable
internal fun PostIdleSections(state: PostCreateUiState, viewModel: PostCreateViewModel) {
    PostFormCard {
        PostSwitchRow(stringResource(R.string.post_fee_setting), state.enableIdlePrice, viewModel::onEnableIdlePrice)
        if (state.enableIdlePrice) {
            PostFormDivider()
            PostFormField(
                label = stringResource(R.string.extra_price),
                value = state.idlePrice,
                onValueChange = viewModel::onIdlePrice,
                placeholder = stringResource(R.string.post_price_hint),
                singleLine = true,
                keyboardType = KeyboardType.Decimal,
                suffix = stringResource(R.string.unit_yuan),
            )
            PostFormDivider()
            PostSwitchRow(stringResource(R.string.post_negotiable), state.idleNegotiable, viewModel::onIdleNegotiable)
        }
    }
}
