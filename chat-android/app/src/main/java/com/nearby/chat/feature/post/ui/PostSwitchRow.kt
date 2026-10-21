package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun PostSwitchRow(
    title: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    val dismissInput = rememberDismissInput()
    val emitChange: (Boolean) -> Unit = { next ->
        dismissInput()
        onCheckedChange(next)
    }
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = { emitChange(!checked) },
            )
            .padding(horizontal = AppSpacing.L, vertical = PostFormRowVPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            title,
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            modifier = Modifier.weight(1f),
        )
        PostSwitch(checked = checked, onCheckedChange = emitChange)
    }
}
