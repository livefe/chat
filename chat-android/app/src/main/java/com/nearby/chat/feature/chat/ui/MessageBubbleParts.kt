package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.sizeIn
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.LineHeightStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val ImageMaxSize = 220.dp
private val HintLineHeight = 17.sp
private val SenderNameLineHeight = 13.sp

@Composable
internal fun CenterHint(text: String) {
    Text(
        text = text,
        color = AppColors.TextMuted,
        fontSize = AppType.Caption,
        lineHeight = HintLineHeight,
        textAlign = TextAlign.Center,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L, vertical = 6.dp),
    )
}

@Composable
internal fun SenderName(name: String) {
    Text(
        text = name,
        fontSize = AppType.Micro,
        lineHeight = SenderNameLineHeight,
        fontWeight = FontWeight.Medium,
        color = AppColors.TextSecondary,
        maxLines = 1,
        overflow = TextOverflow.Ellipsis,
        style = TextStyle(
            lineHeightStyle = LineHeightStyle(
                alignment = LineHeightStyle.Alignment.Top,
                trim = LineHeightStyle.Trim.Both,
            ),
        ),
        modifier = Modifier.padding(horizontal = AppSpacing.Xxs),
    )
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
internal fun ImageMessage(
    url: String,
    failed: Boolean,
    onPreview: (String) -> Unit,
    onLongClick: (() -> Unit)?,
) {
    AsyncImage(
        model = url,
        contentDescription = stringResource(R.string.upload_kind_image),
        contentScale = ContentScale.Fit,
        modifier = Modifier
            .sizeIn(maxWidth = ImageMaxSize, maxHeight = ImageMaxSize)
            .clip(AppShapes.Bubble)
            .background(AppColors.PageBg)
            .then(
                if (url.isNotBlank() || onLongClick != null) {
                    Modifier.combinedClickable(
                        interactionSource = remember { MutableInteractionSource() },
                        indication = null,
                        onClick = { if (url.isNotBlank()) onPreview(url) },
                        onLongClick = onLongClick,
                    )
                } else {
                    Modifier
                },
            ),
    )
    if (failed) {
        SendFailedLabel()
    }
}
