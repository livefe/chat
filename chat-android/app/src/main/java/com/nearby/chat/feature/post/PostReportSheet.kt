package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun PostReportSheet(
    onReport: () -> Unit,
    onDismiss: () -> Unit,
) {
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
    ) {
        Text(
            stringResource(R.string.post_report_action),
            color = AppColors.Danger,
            fontSize = AppSheetTokens.ItemSize,
            fontWeight = FontWeight.Medium,
            textAlign = TextAlign.Center,
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onReport)
                .padding(
                    horizontal = AppSheetTokens.ItemHPad,
                    vertical = AppSheetTokens.ItemVPad,
                ),
        )
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(AppSheetTokens.CancelGap)
                .background(AppColors.PageBg),
        )
        Text(
            stringResource(R.string.common_cancel),
            color = AppColors.TextSecondary,
            fontSize = AppSheetTokens.ItemSize,
            textAlign = TextAlign.Center,
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onDismiss)
                .padding(vertical = AppSheetTokens.ItemVPad),
        )
    }
}
