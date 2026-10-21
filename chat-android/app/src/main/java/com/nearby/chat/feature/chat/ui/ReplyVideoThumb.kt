package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.PlayCircle
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.theme.AppColors

internal val ReplyThumbSize = 28.dp
internal val ReplyThumbShape = RoundedCornerShape(3.dp)
private val ReplyThumbPlayIconSize = 16.dp

@Composable
internal fun ReplyVideoThumb(
    url: String,
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
) {
    val context = LocalContext.current
    val cover = remember(url, context) { videoCoverRequest(context, url) }
    Box(
        modifier = modifier
            .size(ReplyThumbSize)
            .clip(ReplyThumbShape)
            .background(AppColors.PageBg)
            .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier),
        contentAlignment = Alignment.Center,
    ) {
        AsyncImage(
            model = cover,
            contentDescription = stringResource(R.string.upload_kind_video),
            contentScale = ContentScale.Crop,
            modifier = Modifier.fillMaxSize(),
        )
        Icon(
            Icons.Outlined.PlayCircle,
            contentDescription = null,
            tint = Color.White,
            modifier = Modifier.size(ReplyThumbPlayIconSize),
        )
    }
}
