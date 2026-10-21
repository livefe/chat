package com.nearby.chat.feature.profile

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun UserProfileActionSheet(
    blockedByMe: Boolean,
    onDismiss: () -> Unit,
    onBlock: () -> Unit,
    onUnblock: () -> Unit,
    onReport: () -> Unit,
) {
    AppModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
    ) {
        if (blockedByMe) {
            SheetAction(stringResource(R.string.block_unblock_title), AppColors.TextPrimary, onUnblock)
        } else {
            SheetAction(stringResource(R.string.profile_block), AppColors.Danger, onBlock)
        }
        SheetAction(stringResource(R.string.report_title), AppColors.Danger, onReport)
    }
}

@Composable
private fun SheetAction(text: String, color: Color, onClick: () -> Unit) {
    Text(
        text,
        fontSize = AppSheetTokens.ItemSize,
        fontWeight = FontWeight.Medium,
        color = color,
        textAlign = TextAlign.Center,
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(
                horizontal = AppSheetTokens.ItemHPad,
                vertical = AppSheetTokens.ItemVPad,
            ),
    )
}
