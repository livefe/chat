package com.nearby.chat.core.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun AppCellField(
    label: String,
    value: String,
    modifier: Modifier = Modifier,
    onValueChange: (String) -> Unit = {},
    placeholder: String = "",
    readOnly: Boolean = false,
    keyboardType: KeyboardType = KeyboardType.Text,
    fontSize: TextUnit = AppType.Body,
    rowHeight: Dp = 52.dp,
    trailing: (@Composable () -> Unit)? = null,
) {
    val focusRequester = remember { FocusRequester() }
    val rowClick = remember { MutableInteractionSource() }
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = modifier
            .fillMaxWidth()
            .height(rowHeight)
            .then(if (readOnly) Modifier else Modifier.markTextInputPress())
            .then(
                if (readOnly) {
                    Modifier
                } else {
                    Modifier.clickable(
                        interactionSource = rowClick,
                        indication = null,
                        onClick = { focusRequester.requestFocus() },
                    )
                },
            )
            .padding(horizontal = AppSpacing.L),
    ) {
        Text(
            label,
            fontSize = fontSize,
            color = AppColors.TextPrimary,
            modifier = Modifier.width(72.dp),
        )
        if (readOnly) {
            Text(
                value.ifBlank { placeholder },
                fontSize = fontSize,
                color = if (value.isBlank()) AppColors.TextTertiary else AppColors.TextPrimary,
                modifier = Modifier.weight(1f),
            )
        } else {
            BasicTextField(
                value = value,
                onValueChange = onValueChange,
                singleLine = true,
                readOnly = readOnly,
                keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
                textStyle = TextStyle(
                    color = AppColors.TextPrimary,
                    fontSize = fontSize,
                    lineHeight = (fontSize.value + 6).sp,
                ),
                cursorBrush = SolidColor(AppColors.Brand),
                modifier = Modifier
                    .weight(1f)
                    .imeOnTap()
                    .fillMaxHeight()
                    .focusRequester(focusRequester),
                decorationBox = { inner ->
                    Box(
                        modifier = Modifier.fillMaxWidth().fillMaxHeight(),
                        contentAlignment = Alignment.CenterStart,
                    ) {
                        if (value.isEmpty()) {
                            Text(placeholder, color = AppColors.TextTertiary, fontSize = fontSize)
                        }
                        inner()
                    }
                },
            )
        }
        if (trailing != null) {
            trailing()
        }
    }
}
