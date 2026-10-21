package com.nearby.chat.feature.post

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

@Composable
internal fun RemarkEditDialog(
    nickname: String,
    draft: String,
    maxLength: Int,
    saving: Boolean,
    onDraftChange: (String) -> Unit,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    AlertDialog(
        onDismissRequest = { if (!saving) onDismiss() },
        title = {
            Column {
                Text(
                    stringResource(R.string.signup_export_col_remark),
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                )
                if (nickname.isNotBlank()) {
                    Text(
                        nickname,
                        fontSize = AppType.Meta,
                        color = AppColors.TextTertiary,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.padding(top = AppSpacing.Xs),
                    )
                }
            }
        },
        text = {
            OutlinedTextField(
                value = draft,
                onValueChange = onDraftChange,
                modifier = Modifier.fillMaxWidth().imeOnTap(),
                placeholder = {
                    Text(stringResource(R.string.signup_remark_placeholder), color = AppColors.TextTertiary)
                },
                supportingText = {
                    Text(stringResource(R.string.text_counter, draft.length, maxLength), color = AppColors.TextTertiary)
                },
                minLines = 3,
                maxLines = 5,
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = AppColors.Brand,
                    cursorColor = AppColors.Brand,
                    focusedTextColor = AppColors.TextPrimary,
                    unfocusedTextColor = AppColors.TextPrimary,
                ),
            )
        },
        confirmButton = {
            TextButton(onClick = onConfirm, enabled = !saving) {
                Text(
                    stringResource(if (saving) R.string.post_saving else R.string.common_confirm),
                    color = AppColors.TextPrimary,
                )
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss, enabled = !saving) {
                Text(stringResource(R.string.common_cancel), color = AppColors.TextTertiary)
            }
        },
        containerColor = AppColors.CardBg,
    )
}
