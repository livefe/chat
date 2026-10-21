package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.AttachFile
import androidx.compose.material.icons.outlined.Image
import androidx.compose.material.icons.outlined.LocationOn
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
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val MoreTileSize = 56.dp
private val MoreTileShape = RoundedCornerShape(14.dp)
private val MoreTileIconSize = 26.dp
private val MoreTileLabelGap = 6.dp
private val MorePanelBottomPad = 14.dp

@Composable
internal fun ComposerMorePanel(
    enabled: Boolean,
    canCompose: Boolean,
    showLocation: Boolean,
    showFile: Boolean,
    onPickMedia: () -> Unit,
    onPickLocation: () -> Unit,
    onPickFile: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = AppSpacing.L, end = AppSpacing.L, bottom = MorePanelBottomPad),
        horizontalArrangement = Arrangement.spacedBy(AppSpacing.L),
    ) {
        ComposerMoreTile(
            icon = Icons.Outlined.Image,
            label = stringResource(R.string.composer_album),
            enabled = enabled,
            canCompose = canCompose,
            onClick = onPickMedia,
        )
        if (showLocation) {
            ComposerMoreTile(
                icon = Icons.Outlined.LocationOn,
                label = stringResource(R.string.composer_location),
                enabled = enabled,
                canCompose = canCompose,
                onClick = onPickLocation,
            )
        }
        if (showFile) {
            ComposerMoreTile(
                icon = Icons.Outlined.AttachFile,
                label = stringResource(R.string.chat_file_default_name),
                enabled = enabled,
                canCompose = canCompose,
                onClick = onPickFile,
            )
        }
    }
}

@Composable
private fun ComposerMoreTile(
    icon: ImageVector,
    label: String,
    enabled: Boolean,
    canCompose: Boolean,
    onClick: () -> Unit,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.clickable(enabled = enabled, onClick = onClick),
    ) {
        Box(
            modifier = Modifier
                .size(MoreTileSize)
                .clip(MoreTileShape)
                .background(AppColors.ComposerBg),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                icon,
                contentDescription = null,
                tint = if (canCompose) AppColors.TextPrimary else AppColors.TextMuted,
                modifier = Modifier.size(MoreTileIconSize),
            )
        }
        Spacer(modifier = Modifier.height(MoreTileLabelGap))
        Text(label, fontSize = AppType.Caption, color = AppColors.TextSecondary)
    }
}
