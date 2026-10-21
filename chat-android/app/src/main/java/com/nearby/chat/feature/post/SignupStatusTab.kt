package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val IndicatorWidth = 18.dp
private val IndicatorHeight = 3.dp
private val IndicatorShape = RoundedCornerShape(IndicatorHeight)

@Composable
internal fun SignupStatusTab(
    label: String,
    count: Int,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .clickable(onClick = onClick)
            .padding(top = AppSpacing.M),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Row(
            verticalAlignment = Alignment.Bottom,
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.Xs),
        ) {
            Text(
                label,
                fontSize = AppType.Subhead,
                fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
                color = if (selected) AppColors.TextPrimary else AppColors.TextMuted,
            )
            Text(
                count.toString(),
                fontSize = AppType.Caption,
                color = if (selected) AppColors.Brand else AppColors.TextTertiary,
            )
        }
        Spacer(Modifier.height(AppSpacing.S))
        Box(
            Modifier
                .width(IndicatorWidth)
                .height(IndicatorHeight)
                .clip(IndicatorShape)
                .background(if (selected) AppColors.Brand else Color.Transparent),
        )
    }
}
