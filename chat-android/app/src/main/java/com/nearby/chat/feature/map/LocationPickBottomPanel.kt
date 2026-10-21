package com.nearby.chat.feature.map

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.LocationOn
import androidx.compose.material.icons.outlined.MyLocation
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing

private val PanelRowGap = 10.dp
private val PinBadgeSize = 34.dp
private val PinBadgeShape = RoundedCornerShape(10.dp)
private val ProgressSize = 16.dp
private val ActionShape = RoundedCornerShape(20.dp)

@Composable
fun LocationPickBottomPanel(
    state: LocationPickUiState,
    onUseCurrentLocation: () -> Unit,
    onConfirm: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(AppColors.Surface)
            .navigationBarsPadding()
            .padding(horizontal = AppSpacing.L, vertical = AppSpacing.M),
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(PanelRowGap),
        ) {
            Box(
                modifier = Modifier
                    .size(PinBadgeSize)
                    .clip(PinBadgeShape)
                    .background(AppColors.BrandContainer),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    Icons.Outlined.LocationOn,
                    contentDescription = null,
                    tint = AppColors.Brand,
                    modifier = Modifier.size(AppIconSize.Inline),
                )
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                modifier = Modifier.weight(1f),
            ) {
                if (state.geocoding) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(ProgressSize),
                        strokeWidth = 2.dp,
                        color = AppColors.Brand,
                    )
                }
                Text(
                    state.placeName.ifBlank { stringResource(R.string.location_drag_hint) },
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = AppSpacing.M),
            horizontalArrangement = Arrangement.spacedBy(PanelRowGap),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            OutlinedButton(
                onClick = onUseCurrentLocation,
                enabled = !state.locating,
                shape = ActionShape,
                colors = ButtonDefaults.outlinedButtonColors(contentColor = AppColors.Brand),
            ) {
                Icon(
                    Icons.Outlined.MyLocation,
                    contentDescription = null,
                    modifier = Modifier.size(AppIconSize.Inline),
                )
                Text(
                    stringResource(if (state.locating) R.string.location_locating else R.string.location_current),
                    modifier = Modifier.padding(start = AppSpacing.Xs),
                )
            }
            Button(
                onClick = onConfirm,
                enabled = !state.locating,
                modifier = Modifier.weight(1f),
                shape = ActionShape,
                colors = ButtonDefaults.buttonColors(containerColor = AppColors.Brand),
            ) {
                Text(stringResource(R.string.common_confirm))
            }
        }
    }
}
