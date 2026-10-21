package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Add
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.Keyboard
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.imeOnTap

private val ComposerSideButtonWidth = 32.dp
private val ComposerIconSize = 26.dp
private val ComposerMoreIconSize = 15.dp
private val ComposerMoreBorder = 1.5.dp
private val ComposerInputVPad = 9.dp

@Composable
internal fun ComposerVoiceToggle(
    inVoiceMode: Boolean,
    enabled: Boolean,
    tint: Color,
    onToggle: () -> Unit,
) {
    Box(
        modifier = Modifier
            .width(ComposerSideButtonWidth)
            .height(ComposerControlHeight)
            .clip(CircleShape)
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
            ) {
                if (!enabled) return@clickable
                onToggle()
            },
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = if (inVoiceMode) Icons.Outlined.Keyboard else ComposerMicOutlined,
            contentDescription = stringResource(
                if (inVoiceMode) R.string.composer_switch_keyboard else R.string.composer_switch_voice,
            ),
            tint = tint,
            modifier = Modifier.size(ComposerIconSize),
        )
    }
}

@Composable
internal fun ComposerTextInput(
    input: String,
    placeholder: String,
    maxHeight: Dp,
    onValueChange: (String) -> Unit,
    onFocused: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val emptyInput = input.isEmpty()
    BasicTextField(
        value = input,
        onValueChange = onValueChange,
        singleLine = false,
        textStyle = TextStyle(
            fontSize = AppType.Body,
            color = AppColors.TextPrimary,
            lineHeight = ComposerLineHeight,
        ),
        cursorBrush = SolidColor(AppColors.Brand),
        modifier = modifier
            .imeOnTap()
            .then(
                if (emptyInput) {
                    Modifier.height(ComposerControlHeight)
                } else {
                    Modifier.heightIn(min = ComposerControlHeight, max = maxHeight)
                },
            )
            .clip(AppShapes.Composer)
            .background(AppColors.ComposerBg)
            .onFocusChanged { if (it.isFocused) onFocused() }
            .padding(horizontal = AppSpacing.M, vertical = ComposerInputVPad),
        decorationBox = { inner ->
            Box(
                modifier = Modifier.fillMaxWidth(),
                contentAlignment = Alignment.CenterStart,
            ) {
                if (emptyInput) {
                    Text(
                        text = placeholder,
                        color = AppColors.TextMuted,
                        fontSize = AppType.Body,
                        lineHeight = ComposerLineHeight,
                        maxLines = 1,
                    )
                }
                inner()
            }
        },
    )
}

@Composable
internal fun ComposerMoreToggle(
    moreOpen: Boolean,
    enabled: Boolean,
    tint: Color,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .width(ComposerSideButtonWidth)
            .height(ComposerControlHeight)
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Box(
            modifier = Modifier
                .size(ComposerIconSize)
                .clip(CircleShape)
                .border(ComposerMoreBorder, AppColors.ComposerIcon, CircleShape),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                imageVector = if (moreOpen) Icons.Outlined.Close else Icons.Outlined.Add,
                contentDescription = stringResource(
                    if (moreOpen) R.string.composer_collapse else R.string.composer_more,
                ),
                tint = tint,
                modifier = Modifier.size(ComposerMoreIconSize),
            )
        }
    }
}
