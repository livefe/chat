package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

@Composable
fun PostFormField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    singleLine: Boolean,
    minLines: Int = 1,
    readOnly: Boolean = false,
    keyboardType: KeyboardType = KeyboardType.Text,
    suffix: String? = null,
    textAlign: TextAlign = TextAlign.Start,
) {
    val fieldInput: @Composable (Modifier) -> Unit = { inputModifier ->
        if (readOnly) {
            Text(
                value.ifBlank { placeholder },
                fontSize = AppType.Body,
                color = if (value.isBlank()) AppColors.TextTertiary else AppColors.TextPrimary,
                textAlign = textAlign,
                modifier = inputModifier,
            )
        } else {
            BasicTextField(
                value = value,
                onValueChange = onValueChange,
                singleLine = singleLine,
                minLines = minLines,
                readOnly = readOnly,
                keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
                textStyle = TextStyle(
                    color = AppColors.TextPrimary,
                    fontSize = AppType.Body,
                    lineHeight = 20.sp,
                    textAlign = textAlign,
                ),
                cursorBrush = SolidColor(AppColors.Brand),
                modifier = inputModifier.imeOnTap(),
                decorationBox = { inner ->
                    Box(modifier = Modifier.fillMaxWidth()) {
                        if (value.isEmpty()) {
                            Text(
                                placeholder,
                                color = AppColors.TextTertiary,
                                fontSize = AppType.Body,
                                textAlign = textAlign,
                                modifier = Modifier.fillMaxWidth(),
                            )
                        }
                        inner()
                    }
                },
            )
        }
    }

    if (!singleLine) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = AppSpacing.L, vertical = PostFormRowVPad),
            verticalArrangement = Arrangement.spacedBy(AppSpacing.S),
        ) {
            Text(
                label,
                fontSize = AppType.Body,
                color = AppColors.TextPrimary,
            )
            fieldInput(Modifier.fillMaxWidth())
        }
        return
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L, vertical = PostFormRowVPad),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            label,
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            modifier = Modifier
                .width(PostFormLabelWidth)
                .padding(end = AppSpacing.Xs),
        )
        fieldInput(Modifier.weight(1f))
        if (!suffix.isNullOrBlank()) {
            Text(
                suffix,
                color = AppColors.TextTertiary,
                fontSize = AppType.Body,
                modifier = Modifier.padding(start = AppSpacing.S),
            )
        }
    }
}

internal val PostFormLabelWidth = 88.dp
internal val PostFormRowVPad = 14.dp
