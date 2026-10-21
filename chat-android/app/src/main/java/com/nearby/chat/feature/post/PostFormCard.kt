package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.HorizontalDivider
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing

@Composable
internal fun PostFormCard(content: @Composable () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(top = AppSpacing.M)
            .clip(AppShapes.CardLg)
            .background(AppColors.CardBg),
        content = { content() },
    )
}

@Composable
internal fun PostFormDivider() {
    HorizontalDivider(
        color = AppColors.Border,
        modifier = Modifier.padding(start = AppSpacing.L),
    )
}
