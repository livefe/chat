package com.nearby.chat.feature.forward

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

private val SearchShape = RoundedCornerShape(20.dp)
private val SearchInnerHorizontal = 14.dp
private val SearchInnerVertical = 10.dp

@Composable
internal fun ForwardSearchField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    enabled: Boolean,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.M, vertical = AppSpacing.S)
            .clip(SearchShape)
            .background(AppColors.PageBg)
            .padding(horizontal = SearchInnerHorizontal, vertical = SearchInnerVertical),
    ) {
        if (value.isEmpty()) {
            Text(placeholder, fontSize = AppType.Secondary, color = AppColors.TextTertiary)
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            enabled = enabled,
            singleLine = true,
            textStyle = TextStyle(fontSize = AppType.Secondary, color = AppColors.TextPrimary),
            cursorBrush = SolidColor(AppColors.Brand),
            modifier = Modifier.fillMaxWidth().imeOnTap(),
        )
    }
}
