package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.PlatformTextStyle
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.LineHeightStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostCategories
import com.nearby.chat.core.post.PostListHighlights
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

private val ThumbSize = 88.dp
private val ThumbShape = RoundedCornerShape(6.dp)

@Composable
fun PostFeedItem(
    post: PostItem,
    modifier: Modifier = Modifier,
    variant: PostFeedVariant = PostFeedVariant.List,
    badgeMode: PostFeedBadgeMode = PostFeedBadgeMode.Owner,
    onClick: (() -> Unit)? = null,
    footerActions: (@Composable () -> Unit)? = null,
) {
    val cover = post.images.firstOrNull()
    if (variant == PostFeedVariant.Waterfall) {
        PostWaterfallCard(
            post = post,
            cover = cover,
            modifier = modifier,
        )
        return
    }

    val listMetaText = listMeta(post)
    val facts = PostListHighlights.groups(post)
    val cornerBadge = when (badgeMode) {
        PostFeedBadgeMode.Joined -> PostCornerBadge.joined(post, facts.phase)
        PostFeedBadgeMode.Owner -> PostCornerBadge.owner(status = post.status, phase = facts.phase)
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.L, vertical = 14.dp),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(IntrinsicSize.Min)
                .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier),
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.M),
            verticalAlignment = Alignment.Top,
        ) {
            if (!cover.isNullOrBlank()) {
                CoverThumb(cover = cover, imageCount = post.images.size)
            } else {
                PlaceholderThumb(category = post.category)
            }

            Column(
                modifier = Modifier
                    .weight(1f)
                    .heightIn(min = ThumbSize)
                    .fillMaxHeight(),
                verticalArrangement = Arrangement.SpaceBetween,
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                    verticalAlignment = Alignment.Top,
                ) {
                    Text(
                        post.title,
                        modifier = Modifier.weight(1f),
                        fontSize = AppType.Title,
                        fontWeight = FontWeight.SemiBold,
                        lineHeight = 22.sp,
                        color = AppColors.TextPrimary,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                    cornerBadge?.let { badge ->
                        Text(
                            badge.text.asString(),
                            fontSize = AppType.Caption,
                            fontWeight = FontWeight.Medium,
                            color = badge.fg,
                            modifier = Modifier
                                .clip(AppShapes.Capsule)
                                .background(badge.bg)
                                .padding(horizontal = AppSpacing.S, vertical = AppSpacing.Xxs),
                        )
                    }
                }
                if (listMetaText != null || footerActions != null) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(AppSpacing.S),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        if (listMetaText != null) {
                            Text(
                                text = listMetaText,
                                modifier = Modifier.weight(1f),
                                fontSize = AppType.Meta,
                                lineHeight = AppType.LineMeta,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        } else {
                            Spacer(modifier = Modifier.weight(1f))
                        }
                        footerActions?.invoke()
                    }
                }
            }
        }
    }
}

@Composable
private fun listMeta(post: PostItem): AnnotatedString? {
    val timeLabel = PostListHighlights.formatListCreatedAt(post.createdAt)
    val placeLabel = post.placeName.trim()
    if (placeLabel.isBlank() && timeLabel.isBlank()) return null
    val separator = stringResource(R.string.meta_separator)
    return buildAnnotatedString {
        if (placeLabel.isNotBlank()) {
            withStyle(SpanStyle(color = AppColors.TextSecondary, fontSize = AppType.Meta)) {
                append(placeLabel)
            }
        }
        if (placeLabel.isNotBlank() && timeLabel.isNotBlank()) {
            withStyle(SpanStyle(color = AppColors.TextTertiary, fontSize = AppType.Meta)) {
                append(separator)
            }
        }
        if (timeLabel.isNotBlank()) {
            withStyle(SpanStyle(color = AppColors.TextTertiary, fontSize = AppType.Meta)) {
                append(timeLabel)
            }
        }
    }
}

@Composable
private fun CoverThumb(cover: String, imageCount: Int) {
    Box(
        modifier = Modifier
            .size(ThumbSize)
            .clip(ThumbShape)
            .background(AppColors.ThumbPlaceholder),
    ) {
        AsyncImage(
            model = MediaUrl.resolve(cover),
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = Modifier.fillMaxSize(),
        )
        if (imageCount > 1) {
            Box(
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(AppSpacing.Xs)
                    .size(20.dp)
                    .clip(CircleShape)
                    .background(AppColors.CountScrim),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = imageCount.toString(),
                    color = Color.White,
                    fontSize = AppType.Micro,
                    lineHeight = AppType.Micro,
                    fontWeight = FontWeight.Medium,
                    textAlign = TextAlign.Center,
                    style = TextStyle(
                        platformStyle = PlatformTextStyle(includeFontPadding = false),
                        lineHeightStyle = LineHeightStyle(
                            alignment = LineHeightStyle.Alignment.Center,
                            trim = LineHeightStyle.Trim.Both,
                        ),
                    ),
                )
            }
        }
    }
}

@Composable
private fun PlaceholderThumb(category: String) {
    Box(
        modifier = Modifier
            .size(ThumbSize)
            .clip(ThumbShape)
            .background(AppColors.ThumbPlaceholder),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = stringResource(PostCategories.label(category)).take(1)
                .ifBlank { stringResource(R.string.post_cover_placeholder) },
            fontSize = AppType.Profile,
            fontWeight = FontWeight.SemiBold,
            color = AppColors.BrandDark,
        )
    }
}
