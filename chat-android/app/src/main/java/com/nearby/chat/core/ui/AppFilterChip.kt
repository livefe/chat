package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun AppFilterChip(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val bg = if (selected) AppColors.BrandContainer else AppColors.CardBg
    val border = if (selected) AppColors.Brand else AppColors.BorderSoft
    val fg = if (selected) AppColors.BrandDark else AppColors.TextSecondary
    Box(
        modifier = modifier
            .clip(AppShapes.Chip)
            .border(1.dp, border, AppShapes.Chip)
            .background(bg, AppShapes.Chip)
            .clickable(onClick = onClick)
            .padding(horizontal = AppSpacing.M, vertical = 6.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            color = fg,
            fontSize = AppType.Meta,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Medium,
        )
    }
}
