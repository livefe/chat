package com.nearby.chat.feature.chat

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import com.nearby.chat.R
import com.nearby.chat.core.model.RoomPlaceType
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun RoomPlaceTypeSheet(
    selectedType: Int?,
    onSelect: (Int) -> Unit,
    onDismiss: () -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
    ) {
        RoomPlaceType.options.forEachIndexed { index, option ->
            if (index > 0) {
                HorizontalDivider(color = AppColors.Border)
            }
            val selected = selectedType == option.value
            Text(
                stringResource(option.label),
                fontSize = AppSheetTokens.ItemSize,
                color = if (selected) AppColors.Brand else AppColors.TextPrimary,
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onSelect(option.value) }
                    .padding(vertical = AppSheetTokens.ItemVPad),
            )
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(AppSheetTokens.CancelGap)
                .background(AppColors.PageBg),
        )
        Text(
            stringResource(R.string.common_cancel),
            fontSize = AppSheetTokens.ItemSize,
            color = AppColors.TextSecondary,
            textAlign = TextAlign.Center,
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onDismiss)
                .padding(vertical = AppSheetTokens.ItemVPad),
        )
    }
}
