package com.nearby.chat.feature.chat

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Check
import androidx.compose.material.icons.outlined.MyLocation
import androidx.compose.material.icons.outlined.Place
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun NearbyLocationSheet(
    gpsSelected: Boolean,
    mapSelected: Boolean,
    onDismiss: () -> Unit,
    onRelocate: () -> Unit,
    onMapPick: () -> Unit,
) {
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
    ) {
        LocationSheetItem(
            icon = Icons.Outlined.MyLocation,
            label = stringResource(R.string.nearby_relocate),
            selected = gpsSelected,
            onClick = onRelocate,
        )
        LocationSheetItem(
            icon = Icons.Outlined.Place,
            label = stringResource(R.string.nearby_map_pick),
            selected = mapSelected,
            onClick = onMapPick,
        )
    }
}

@Composable
private fun LocationSheetItem(
    icon: ImageVector,
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(
                horizontal = AppSheetTokens.ItemHPad,
                vertical = AppSheetTokens.ItemVPad,
            ),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            icon,
            contentDescription = null,
            tint = AppColors.TextPrimary,
            modifier = Modifier.size(AppIconSize.Cell),
        )
        Spacer(Modifier.width(AppSpacing.M))
        Text(
            label,
            modifier = Modifier.weight(1f),
            fontSize = AppSheetTokens.ItemSize,
            fontWeight = FontWeight.Medium,
            color = AppColors.TextPrimary,
        )
        if (selected) {
            Icon(
                Icons.Outlined.Check,
                contentDescription = null,
                tint = AppColors.Brand,
                modifier = Modifier.size(AppIconSize.Inline),
            )
        }
    }
}
