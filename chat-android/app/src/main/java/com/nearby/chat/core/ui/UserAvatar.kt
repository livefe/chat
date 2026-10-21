package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.Dp
import coil.compose.AsyncImage
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes

@Composable
fun UserAvatar(
    url: String?,
    size: Dp,
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
) {
    val shapeMod = Modifier
        .size(size)
        .clip(AppShapes.Avatar)
        .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
        .then(modifier)
    if (url.isNullOrBlank()) {
        Box(modifier = shapeMod.background(AppColors.BorderSoft))
    } else {
        AsyncImage(
            model = MediaUrl.resolve(url),
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = shapeMod.background(AppColors.BorderSoft),
        )
    }
}
