package com.nearby.chat.feature.forward

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Link
import androidx.compose.material.icons.outlined.Photo
import androidx.compose.material.icons.outlined.Share
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val BarVerticalPadding = 10.dp
private val ActionLabelGap = 6.dp

@Composable
internal fun ForwardExternalShareBar(
    sharing: Boolean,
    onSaveImage: () -> Unit,
    onCopyLink: () -> Unit,
    onShare: () -> Unit,
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(AppColors.CardBg)
            .navigationBarsPadding()
            .padding(horizontal = AppSpacing.L, vertical = BarVerticalPadding),
    ) {
        Text(
            stringResource(R.string.forward_external),
            fontSize = AppType.Caption,
            color = AppColors.TextTertiary,
        )
        Spacer(Modifier.height(AppSpacing.S))
        Row(modifier = Modifier.fillMaxWidth()) {
            ExternalAction(
                icon = Icons.Outlined.Photo,
                label = stringResource(R.string.forward_save_image),
                enabled = !sharing,
                onClick = onSaveImage,
                modifier = Modifier.weight(1f),
            )
            Spacer(Modifier.width(AppSpacing.S))
            ExternalAction(
                icon = Icons.Outlined.Link,
                label = stringResource(R.string.forward_copy_link),
                enabled = true,
                onClick = onCopyLink,
                modifier = Modifier.weight(1f),
            )
            Spacer(Modifier.width(AppSpacing.S))
            ExternalAction(
                icon = Icons.Outlined.Share,
                label = stringResource(R.string.common_more),
                enabled = !sharing,
                onClick = onShare,
                modifier = Modifier.weight(1f),
            )
        }
    }
}

@Composable
private fun ExternalAction(
    icon: ImageVector,
    label: String,
    enabled: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = modifier
            .clip(AppShapes.Card)
            .background(AppColors.PageBg)
            .clickable(enabled = enabled, onClick = onClick)
            .padding(vertical = BarVerticalPadding),
    ) {
        Icon(icon, contentDescription = null, tint = AppColors.TextSecondary, modifier = Modifier.size(AppIconSize.Cell))
        Spacer(Modifier.height(ActionLabelGap))
        Text(label, fontSize = AppType.Caption, color = AppColors.TextSecondary)
    }
}
