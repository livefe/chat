package com.nearby.chat.core.ui

import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.KeyboardArrowLeft
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBarColors
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppType

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun appTopAppBarColors(containerColor: Color = AppColors.PageBg): TopAppBarColors = TopAppBarDefaults.topAppBarColors(
    containerColor = containerColor,
    scrolledContainerColor = containerColor,
    titleContentColor = AppColors.TextPrimary,
    navigationIconContentColor = AppColors.TextPrimary,
    actionIconContentColor = AppColors.TextPrimary,
)

@Composable
fun AppTopBarTitle(
    text: String,
    modifier: Modifier = Modifier,
    maxLines: Int = 1,
) {
    Text(
        text = text,
        modifier = modifier,
        fontSize = AppType.PageTitle,
        fontWeight = FontWeight.SemiBold,
        color = AppColors.TextPrimary,
        maxLines = maxLines,
        overflow = TextOverflow.Ellipsis,
    )
}

@Composable
fun AppBackIcon(
    modifier: Modifier = Modifier,
    size: Dp = 28.dp,
    contentDescription: String? = stringResource(R.string.common_back),
) {
    Icon(
        imageVector = Icons.AutoMirrored.Filled.KeyboardArrowLeft,
        contentDescription = contentDescription,
        modifier = modifier.size(size),
    )
}

@Composable
fun AppBackIconButton(
    onClick: () -> Unit,
    contentDescription: String? = stringResource(R.string.common_back),
) {
    IconButton(onClick = onClick) {
        AppBackIcon(contentDescription = contentDescription)
    }
}
