package com.nearby.chat.feature.post

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

@Composable
internal fun PostDetailImages(images: List<String>) {
    val urls = remember(images) { images.map { MediaUrl.resolve(it) }.filter { it.isNotBlank() } }
    if (urls.isEmpty()) return
    var previewIndex by remember { mutableStateOf<Int?>(null) }

    if (urls.size == 1) {
        AsyncImage(
            model = urls[0],
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(16f / 10f)
                .clip(AppShapes.Media)
                .background(AppColors.MediaPlaceholder)
                .clickable { previewIndex = 0 },
        )
    } else {
        val pagerState = rememberPagerState(pageCount = { urls.size })
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(4f / 3f)
                .clip(AppShapes.Media)
                .background(AppColors.MediaPlaceholder),
        ) {
            HorizontalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                AsyncImage(
                    model = urls[page],
                    contentDescription = null,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxSize()
                        .clickable { previewIndex = page },
                )
            }
            Text(
                text = stringResource(R.string.image_counter, pagerState.currentPage + 1, urls.size),
                color = Color.White,
                fontSize = AppType.Caption,
                lineHeight = AppType.LineMeta,
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(10.dp)
                    .clip(AppShapes.Capsule)
                    .background(AppColors.MediaScrim)
                    .padding(horizontal = AppSpacing.S, vertical = AppSpacing.Xxs),
            )
        }
    }

    val start = previewIndex
    if (start != null) {
        PostImagePreview(
            urls = urls,
            initialPage = start,
            onDismiss = { previewIndex = null },
        )
    }
}

@Composable
private fun PostImagePreview(
    urls: List<String>,
    initialPage: Int,
    onDismiss: () -> Unit,
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(
            usePlatformDefaultWidth = false,
            decorFitsSystemWindows = false,
        ),
    ) {
        val previewPager = rememberPagerState(initialPage = initialPage, pageCount = { urls.size })
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            HorizontalPager(
                state = previewPager,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .clickable(onClick = onDismiss),
                    contentAlignment = Alignment.Center,
                ) {
                    AsyncImage(
                        model = urls[page],
                        contentDescription = stringResource(R.string.image_preview),
                        contentScale = ContentScale.Fit,
                        modifier = Modifier.fillMaxWidth(),
                    )
                }
            }
            if (urls.size > 1) {
                Text(
                    text = stringResource(R.string.image_counter, previewPager.currentPage + 1, urls.size),
                    color = Color.White,
                    fontSize = AppType.Meta,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .navigationBarsPadding()
                        .padding(bottom = AppSpacing.Xxl)
                        .clip(AppShapes.Capsule)
                        .background(AppColors.MediaScrim)
                        .padding(horizontal = 10.dp, vertical = AppSpacing.Xs),
                )
            }
        }
    }
}
