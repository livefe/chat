package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun ChatLoadMoreHint(modifier: Modifier = Modifier) {
    Text(
        text = stringResource(R.string.chat_load_more_hint),
        color = AppColors.TextMuted,
        fontSize = AppType.Meta,
        textAlign = TextAlign.Center,
        modifier = modifier
            .fillMaxWidth()
            .padding(vertical = AppSpacing.S),
    )
}
