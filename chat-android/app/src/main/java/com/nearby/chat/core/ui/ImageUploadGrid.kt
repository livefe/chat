package com.nearby.chat.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Close
import androidx.compose.material.icons.outlined.PhotoCamera
import androidx.compose.material.icons.outlined.Refresh
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val CellSize = 80.dp
private val CellGap = AppSpacing.S

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun ImageUploadGrid(
    items: List<ImageUploadItem>,
    canAdd: Boolean,
    onAdd: () -> Unit,
    onRemove: (String) -> Unit,
    onRetry: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    FlowRow(
        modifier = modifier,
        horizontalArrangement = Arrangement.spacedBy(CellGap),
        verticalArrangement = Arrangement.spacedBy(CellGap),
    ) {
        items.forEach { item ->
            ImageUploadThumb(
                item = item,
                onRemove = { onRemove(item.id) },
                onRetry = { onRetry(item.id) },
            )
        }
        if (canAdd) {
            ImageUploadAddCell(onClick = onAdd)
        }
    }
}

@Composable
private fun ImageUploadAddCell(onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(CellSize)
            .clip(RectangleShape)
            .background(AppColors.PageBg)
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            Icons.Outlined.PhotoCamera,
            contentDescription = stringResource(R.string.image_add),
            tint = AppColors.TextTertiary,
            modifier = Modifier.size(24.dp),
        )
    }
}

@Composable
private fun ImageUploadThumb(
    item: ImageUploadItem,
    onRemove: () -> Unit,
    onRetry: () -> Unit,
) {
    val preview = item.url.takeIf { it.isNotBlank() }?.let { MediaUrl.resolve(it) } ?: item.localUri
    Box(Modifier.size(CellSize)) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .clip(RectangleShape)
                .background(AppColors.PageBg),
        ) {
            AsyncImage(
                model = preview,
                contentDescription = null,
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
            )
            when {
                item.uploading -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(AppColors.UploadScrim),
                        contentAlignment = Alignment.Center,
                    ) {
                        CircularProgressIndicator(
                            color = Color.White,
                            strokeWidth = 2.dp,
                            modifier = Modifier.size(20.dp),
                        )
                    }
                }

                item.failed -> {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(AppColors.UploadFailScrim)
                            .clickable(onClick = onRetry),
                        contentAlignment = Alignment.Center,
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Icon(
                                Icons.Outlined.Refresh,
                                contentDescription = stringResource(R.string.common_retry),
                                tint = Color.White,
                                modifier = Modifier.size(AppIconSize.Inline),
                            )
                            Text(
                                stringResource(R.string.common_retry),
                                color = Color.White,
                                fontSize = AppType.Micro,
                                modifier = Modifier.padding(top = AppSpacing.Xxs),
                            )
                        }
                    }
                }
            }
        }
        Box(
            modifier = Modifier
                .align(Alignment.TopEnd)
                .size(18.dp)
                .background(AppColors.UploadDeleteBg)
                .clickable(onClick = onRemove),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                Icons.Outlined.Close,
                contentDescription = stringResource(R.string.common_delete),
                tint = Color.White,
                modifier = Modifier.size(10.dp),
            )
        }
    }
}
