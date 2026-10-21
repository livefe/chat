package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.ChatBubbleOutline
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.UserAvatar

private const val PREVIEW_MAX_LINES = 6
private val WHITESPACE = Regex("\\s+")

@Composable
internal fun PostWaterfallCard(
    post: PostItem,
    cover: String?,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .clip(AppShapes.Media)
            .background(AppColors.CardBg),
    ) {
        Column(modifier = Modifier.fillMaxWidth()) {
            if (!cover.isNullOrBlank()) {
                WaterfallCover(cover = cover, label = "waterfallCoverAlpha")
            } else {
                val preview = post.content.trim().replace(WHITESPACE, " ").ifBlank { post.title.trim() }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(WATERFALL_RATIO_DEFAULT)
                        .background(Brush.linearGradient(colors = WaterfallPreviewGradient)),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = 18.dp, vertical = AppSpacing.Xl),
                        verticalArrangement = Arrangement.Center,
                        horizontalAlignment = Alignment.CenterHorizontally,
                    ) {
                        if (preview.isNotBlank()) {
                            Text(
                                text = preview,
                                fontSize = AppType.Body,
                                lineHeight = 25.sp,
                                fontWeight = FontWeight.Medium,
                                letterSpacing = 0.15.sp,
                                color = WaterfallPreviewText,
                                maxLines = PREVIEW_MAX_LINES,
                                overflow = TextOverflow.Ellipsis,
                                textAlign = TextAlign.Center,
                                modifier = Modifier.fillMaxWidth(),
                            )
                        }
                    }
                }
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = AppSpacing.S),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    post.title,
                    fontSize = AppType.Meta,
                    fontWeight = FontWeight.SemiBold,
                    lineHeight = AppType.LineMeta,
                    color = AppColors.TextPrimary,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                WaterfallFooter(post)
            }
        }
    }
}

@Composable
private fun WaterfallFooter(post: PostItem) {
    val publisher = post.publisher
    val discussCount = post.discussMessageCount
    if (publisher == null && discussCount <= 0) return
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
    ) {
        if (publisher != null) {
            Row(
                modifier = Modifier.weight(1f),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                UserAvatar(
                    url = publisher.avatarUrl,
                    size = 18.dp,
                )
                Text(
                    publisher.nickname,
                    fontSize = AppType.Micro,
                    lineHeight = 14.sp,
                    color = AppColors.TextTertiary,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f),
                )
            }
        } else {
            Spacer(modifier = Modifier.weight(1f))
        }
        if (discussCount > 0) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(3.dp),
            ) {
                Icon(
                    imageVector = Icons.Outlined.ChatBubbleOutline,
                    contentDescription = null,
                    tint = AppColors.TextTertiary,
                    modifier = Modifier.size(15.dp),
                )
                Text(
                    discussCount.toString(),
                    fontSize = AppType.Caption,
                    lineHeight = 15.sp,
                    color = AppColors.TextTertiary,
                )
            }
        }
    }
}
