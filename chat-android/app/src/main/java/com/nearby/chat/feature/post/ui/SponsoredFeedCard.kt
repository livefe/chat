package com.nearby.chat.feature.post.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.nearby.chat.R
import com.nearby.chat.core.model.SponsoredItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private const val TITLE_PREVIEW_MAX_LINES = 5
private val AdBadgeText = AppColors.AdBadgeText
private val AdBadgeBorder = AppColors.AdBadgeBorder
private val AdBadgeShape = RoundedCornerShape(2.dp)

@Composable
fun SponsoredFeedCard(
    item: SponsoredItem,
    onImpression: (SponsoredItem) -> Unit,
    onClick: (SponsoredItem) -> Unit,
    modifier: Modifier = Modifier,
) {
    LaunchedEffect(item.impressionToken) {
        if (item.impressionToken.isNotBlank()) {
            onImpression(item)
        }
    }

    val cover = item.imageUrl.trim()
    val advertiserLabel = item.advertiserName.trim().ifBlank { stringResource(R.string.ad_advertiser_default) }
    val adLabel = stringResource(R.string.ad_badge_default)
    val adBadge = item.badge.trim().ifBlank { adLabel }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .clip(AppShapes.Media)
            .background(AppColors.CardBg)
            .clickable { onClick(item) },
    ) {
        Column(modifier = Modifier.fillMaxWidth()) {
            if (cover.isNotBlank()) {
                WaterfallCover(cover = cover, label = "sponsoredCoverAlpha")
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(WATERFALL_RATIO_DEFAULT)
                        .background(Brush.linearGradient(colors = WaterfallPreviewGradient)),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = item.title,
                        fontSize = AppType.Subhead,
                        lineHeight = 22.sp,
                        fontWeight = FontWeight.Medium,
                        color = WaterfallPreviewText,
                        maxLines = TITLE_PREVIEW_MAX_LINES,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.padding(horizontal = AppSpacing.L),
                    )
                }
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = AppSpacing.S),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    item.title,
                    fontSize = AppType.Meta,
                    fontWeight = FontWeight.SemiBold,
                    lineHeight = AppType.LineMeta,
                    color = AppColors.TextPrimary,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween,
                ) {
                    Text(
                        advertiserLabel,
                        fontSize = AppType.Micro,
                        lineHeight = 14.sp,
                        color = AppColors.TextTertiary,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier
                            .weight(1f)
                            .padding(end = AppSpacing.S),
                    )
                    Text(
                        adBadge,
                        fontSize = 10.sp,
                        lineHeight = 12.sp,
                        fontWeight = FontWeight.SemiBold,
                        letterSpacing = 0.6.sp,
                        color = AdBadgeText,
                        maxLines = 1,
                        modifier = Modifier
                            .semantics { contentDescription = adLabel }
                            .border(1.dp, AdBadgeBorder, AdBadgeShape)
                            .padding(horizontal = AppSpacing.Xs, vertical = AppSpacing.Xxs),
                    )
                }
            }
        }
    }
}
