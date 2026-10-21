package com.nearby.chat.feature.forward

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppType

private val TabLabelPadding = 10.dp
private val IndicatorWidth = 28.dp
private val IndicatorHeight = 2.dp

@Composable
internal fun ForwardTabItem(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.clickable(onClick = onClick),
    ) {
        Text(
            label,
            fontSize = AppType.Subhead,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            color = if (selected) AppColors.TextPrimary else AppColors.TextSecondary,
            modifier = Modifier.padding(vertical = TabLabelPadding),
        )
        Box(
            modifier = Modifier
                .width(IndicatorWidth)
                .height(IndicatorHeight)
                .background(if (selected) AppColors.Brand else Color.Transparent),
        )
    }
}
