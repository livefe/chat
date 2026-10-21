package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.VerticalDivider
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val DialogHPad = 40.dp
private val DialogActionHeight = 48.dp
private val DialogHairline = 0.5.dp
private val DialogMessageLineHeight = 22.sp

@Composable
fun AppConfirmDialog(
    title: String,
    message: String,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
    confirmText: String = stringResource(R.string.common_ok),
    cancelText: String = stringResource(R.string.common_cancel),
    danger: Boolean = false,
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false),
    ) {
        Column(
            modifier = Modifier
                .padding(horizontal = DialogHPad)
                .fillMaxWidth()
                .clip(AppShapes.CardLg)
                .background(AppColors.CardBg),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        start = AppSpacing.Xl,
                        end = AppSpacing.Xl,
                        top = AppSpacing.Xxl,
                        bottom = AppSpacing.Xl,
                    ),
            ) {
                Text(
                    title,
                    fontSize = AppType.PageTitle,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                    textAlign = TextAlign.Start,
                    modifier = Modifier.fillMaxWidth(),
                )
                Text(
                    message,
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                    lineHeight = DialogMessageLineHeight,
                    textAlign = TextAlign.Start,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = AppSpacing.S),
                )
            }

            HorizontalDivider(color = AppColors.BorderSoft, thickness = DialogHairline)
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(DialogActionHeight),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .height(DialogActionHeight)
                        .clickable(onClick = onDismiss),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        cancelText,
                        fontSize = AppSheetTokens.ActionSize,
                        color = AppColors.TextPrimary,
                    )
                }
                VerticalDivider(
                    modifier = Modifier
                        .height(DialogActionHeight)
                        .width(DialogHairline),
                    color = AppColors.BorderSoft,
                )
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .height(DialogActionHeight)
                        .clickable(onClick = onConfirm),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        confirmText,
                        fontSize = AppSheetTokens.ActionSize,
                        fontWeight = FontWeight.Medium,
                        color = if (danger) AppColors.Danger else AppColors.Brand,
                    )
                }
            }
        }
    }
}
