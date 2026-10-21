package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.MoreHoriz
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.ui.AppModalBottomSheet
import com.nearby.chat.core.ui.AppSheetTokens

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostListActionBar(
    actions: List<PostListAction>,
    modifier: Modifier = Modifier,
) {
    if (actions.isEmpty()) return

    var sheetVisible by remember { mutableStateOf(false) }
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    Box(
        contentAlignment = Alignment.Center,
        modifier = modifier
            .size(width = 28.dp, height = 18.dp)
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = { sheetVisible = true },
            ),
    ) {
        Icon(
            imageVector = Icons.Outlined.MoreHoriz,
            contentDescription = stringResource(R.string.common_manage),
            tint = AppColors.TextTertiary,
            modifier = Modifier.size(AppIconSize.Inline),
        )
    }

    if (sheetVisible) {
        AppModalBottomSheet(
            onDismissRequest = { sheetVisible = false },
            sheetState = sheetState,
        ) {
            actions.forEachIndexed { index, action ->
                if (index > 0) {
                    HorizontalDivider(color = AppColors.Border)
                }
                Text(
                    text = action.label,
                    fontSize = AppSheetTokens.ItemSize,
                    fontWeight = FontWeight.Medium,
                    color = if (action.danger) AppColors.Danger else AppColors.TextPrimary,
                    textAlign = TextAlign.Center,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(
                            enabled = action.enabled,
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                            onClick = {
                                sheetVisible = false
                                action.onClick()
                            },
                        )
                        .padding(
                            horizontal = AppSheetTokens.ItemHPad,
                            vertical = AppSheetTokens.ItemVPad,
                        ),
                )
            }
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(AppSheetTokens.CancelGap)
                    .background(AppColors.PageBg),
            )
            Text(
                text = stringResource(R.string.common_cancel),
                fontSize = AppSheetTokens.ItemSize,
                color = AppColors.TextSecondary,
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { sheetVisible = false }
                    .padding(vertical = AppSheetTokens.ItemVPad),
            )
        }
    }
}
