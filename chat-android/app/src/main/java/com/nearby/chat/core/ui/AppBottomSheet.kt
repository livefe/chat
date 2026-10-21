package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.SheetState
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing

@Composable
fun AppSheetHandle(modifier: Modifier = Modifier) {
    Box(
        modifier
            .padding(bottom = AppSheetTokens.HandleGap)
            .size(width = AppSheetTokens.HandleWidth, height = AppSheetTokens.HandleHeight)
            .clip(AppShapes.Capsule)
            .background(AppColors.SheetHandle),
    )
}

@Composable
fun AppSheetTitle(
    text: String,
    modifier: Modifier = Modifier,
) {
    Text(
        text = text,
        fontSize = AppSheetTokens.TitleSize,
        fontWeight = FontWeight.SemiBold,
        color = AppColors.TextPrimary,
        modifier = modifier.padding(horizontal = AppSheetTokens.TitleHPad, vertical = AppSpacing.Xs),
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AppModalBottomSheet(
    onDismissRequest: () -> Unit,
    modifier: Modifier = Modifier,
    sheetState: SheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
    showHandle: Boolean = true,
    containerColor: Color = AppColors.CardBg,
    content: @Composable ColumnScope.() -> Unit,
) {
    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
        sheetState = sheetState,
        containerColor = containerColor,
        shape = AppShapes.Sheet,
        dragHandle = null,
        contentWindowInsets = { WindowInsets(0, 0, 0, 0) },
    ) {
        ProvideAppTextToolbar {
            Column(
                modifier = modifier
                    .fillMaxWidth()
                    .navigationBarsPadding()
                    .padding(
                        top = if (showHandle) AppSheetTokens.PadTop else 0.dp,
                        bottom = AppSheetTokens.PadBottom,
                    ),
            ) {
                if (showHandle) {
                    AppSheetHandle(modifier = Modifier.align(Alignment.CenterHorizontally))
                }
                content()
            }
        }
    }
}
