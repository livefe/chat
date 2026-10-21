package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun PostClickRow(
    label: String,
    value: String,
    placeholder: String,
    onClick: () -> Unit,
    compactBottom: Boolean = false,
) {
    val dismissInput = rememberDismissInput()
    val isPlaceholder = value.isBlank()
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable {
                dismissInput()
                onClick()
            }
            .padding(
                start = AppSpacing.L,
                end = AppSpacing.L,
                top = PostFormRowVPad,
                bottom = if (compactBottom) AppSpacing.S else PostFormRowVPad,
            ),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            label,
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            modifier = Modifier
                .width(PostFormLabelWidth)
                .padding(end = AppSpacing.Xs),
        )
        Text(
            value.ifBlank { placeholder },
            fontSize = AppType.Body,
            color = if (isPlaceholder) AppColors.TextTertiary else AppColors.TextPrimary,
            textAlign = TextAlign.End,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f),
        )
        Icon(
            Icons.AutoMirrored.Outlined.KeyboardArrowRight,
            contentDescription = null,
            tint = AppColors.TextTertiary,
            modifier = Modifier.size(AppIconSize.Inline),
        )
    }
}
