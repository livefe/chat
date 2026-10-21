package com.nearby.chat.core.ui

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
fun ComposerSendButton(
    onClick: () -> Unit,
    enabled: Boolean,
    modifier: Modifier = Modifier,
) {
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier.height(40.dp),
        shape = AppShapes.Composer,
        colors = ButtonDefaults.buttonColors(
            containerColor = AppColors.Brand,
            contentColor = Color.White,
            disabledContainerColor = AppColors.Brand.copy(alpha = 0.45f),
            disabledContentColor = Color.White.copy(alpha = 0.45f),
        ),
        contentPadding = PaddingValues(horizontal = AppSpacing.L),
        elevation = ButtonDefaults.buttonElevation(
            defaultElevation = 0.dp,
            pressedElevation = 0.dp,
            disabledElevation = 0.dp,
        ),
    ) {
        Text(
            stringResource(R.string.common_send),
            fontSize = AppType.Body,
            fontWeight = FontWeight.SemiBold,
            lineHeight = AppType.Body,
        )
    }
}

@Composable
fun AppPrimaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    fillMaxWidth: Boolean = true,
    compact: Boolean = false,
    fontSize: TextUnit? = null,
) {
    val resolvedSize = fontSize ?: if (compact) AppType.Secondary else AppType.Body
    Button(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier
            .then(if (fillMaxWidth && !compact) Modifier.fillMaxWidth() else Modifier)
            .then(if (compact) Modifier.heightIn(min = 34.dp) else Modifier.heightIn(min = 44.dp)),
        shape = AppShapes.Pill,
        colors = ButtonDefaults.buttonColors(
            containerColor = AppColors.Brand,
            contentColor = Color.White,
            disabledContainerColor = AppColors.Brand.copy(alpha = 0.35f),
            disabledContentColor = Color.White.copy(alpha = 0.9f),
        ),
        contentPadding = if (compact) {
            PaddingValues(horizontal = 18.dp, vertical = 10.dp)
        } else {
            PaddingValues(horizontal = AppSpacing.L, vertical = AppSpacing.M)
        },
        elevation = ButtonDefaults.buttonElevation(defaultElevation = 0.dp, pressedElevation = 0.dp),
    ) {
        Text(
            text,
            fontSize = resolvedSize,
            fontWeight = FontWeight.SemiBold,
            lineHeight = resolvedSize,
        )
    }
}

@Composable
fun AppSecondaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    fillMaxWidth: Boolean = true,
    compact: Boolean = false,
) {
    OutlinedButton(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier
            .then(if (fillMaxWidth && !compact) Modifier.fillMaxWidth() else Modifier)
            .then(if (compact) Modifier.heightIn(min = 34.dp) else Modifier.heightIn(min = 44.dp)),
        shape = AppShapes.Pill,
        border = BorderStroke(1.dp, AppColors.BorderSoft),
        colors = ButtonDefaults.outlinedButtonColors(
            contentColor = AppColors.TextPrimary,
            disabledContentColor = AppColors.TextTertiary,
        ),
        contentPadding = if (compact) {
            PaddingValues(horizontal = 18.dp, vertical = 10.dp)
        } else {
            PaddingValues(horizontal = AppSpacing.L, vertical = AppSpacing.M)
        },
    ) {
        Text(
            text,
            fontSize = if (compact) AppType.Secondary else AppType.Body,
            fontWeight = FontWeight.Medium,
            lineHeight = if (compact) AppType.Secondary else AppType.Body,
        )
    }
}

@Composable
fun AppPlainPrimaryButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    fontSize: TextUnit = AppType.Secondary,
) {
    OutlinedButton(
        onClick = onClick,
        enabled = enabled,
        shape = AppShapes.Tag,
        border = BorderStroke(
            1.dp,
            if (enabled) AppColors.Brand else AppColors.Brand.copy(alpha = 0.35f),
        ),
        colors = ButtonDefaults.outlinedButtonColors(
            contentColor = AppColors.Brand,
            disabledContentColor = AppColors.Brand.copy(alpha = 0.35f),
        ),
        contentPadding = PaddingValues(horizontal = 10.dp, vertical = 0.dp),
        modifier = modifier.height(30.dp),
    ) {
        Text(text, fontSize = fontSize)
    }
}
