package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppType

private val TimeDividerBg = AppColors.NeutralLine
private val TimeDividerShape = AppShapes.Capsule
private val TimeDividerVPad = 10.dp
private val TimeDividerChipHPad = 10.dp
private val TimeDividerChipVPad = 3.dp

@Composable
fun ChatTimeDivider(ts: Long, modifier: Modifier = Modifier) {
    val text = formatChatBubbleTime(LocalContext.current, ts)
    if (text.isBlank()) return
    Box(
        modifier = modifier
            .fillMaxWidth()
            .padding(vertical = TimeDividerVPad),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = text,
            fontSize = AppType.Micro,
            color = AppColors.TextMuted,
            modifier = Modifier
                .clip(TimeDividerShape)
                .background(TimeDividerBg)
                .padding(horizontal = TimeDividerChipHPad, vertical = TimeDividerChipVPad),
        )
    }
}
