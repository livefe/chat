package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.LocationOn
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLocationContent
import com.nearby.chat.core.chat.ChatPostContent
import com.nearby.chat.core.chat.chatPostTitle
import com.nearby.chat.core.chat.displayName
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppIconSize
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

private val CardHPad = AppSpacing.M
private val CardVPad = 10.dp
private val CardGap = 10.dp
private val PostCoverSize = 48.dp
private val PostCoverShape = RoundedCornerShape(6.dp)
private val FileGlyphSize = 20.sp

@Composable
private fun CardRow(
    minWidth: Dp,
    maxWidth: Dp,
    content: @Composable RowScope.() -> Unit,
) {
    Row(
        modifier = Modifier
            .widthIn(min = minWidth, max = maxWidth)
            .padding(horizontal = CardHPad, vertical = CardVPad),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(CardGap),
        content = content,
    )
}

@Composable
private fun CardHint(text: String) {
    Text(text = text, fontSize = AppType.Micro, color = AppColors.TextSecondary)
}

@Composable
internal fun PostCardContent(post: ChatPostContent) {
    CardRow(minWidth = 196.dp, maxWidth = 240.dp) {
        val coverUrl = MediaUrl.resolve(post.cover)
        if (coverUrl.isNotBlank()) {
            AsyncImage(
                model = coverUrl,
                contentDescription = null,
                contentScale = ContentScale.Crop,
                modifier = Modifier
                    .size(PostCoverSize)
                    .clip(PostCoverShape)
                    .background(AppColors.PageBg),
            )
        } else {
            Box(
                modifier = Modifier
                    .size(PostCoverSize)
                    .clip(PostCoverShape)
                    .background(AppColors.PageBg),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = stringResource(R.string.chat_post_card_mark),
                    fontSize = AppType.Meta,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.Brand,
                )
            }
        }
        Column(modifier = Modifier.weight(1f, fill = false)) {
            Text(
                text = chatPostTitle(post).asString(),
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            CardHint(stringResource(R.string.chat_post_card_hint))
        }
    }
}

@Composable
internal fun LocationCardContent(location: ChatLocationContent) {
    CardRow(minWidth = 168.dp, maxWidth = 220.dp) {
        Icon(
            Icons.Outlined.LocationOn,
            contentDescription = null,
            tint = AppColors.Brand,
            modifier = Modifier.size(AppIconSize.Cell),
        )
        Column(modifier = Modifier.weight(1f, fill = false)) {
            Text(
                text = location.displayName(),
                fontSize = AppType.Body,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            CardHint(stringResource(R.string.chat_location_card_hint))
        }
    }
}

@Composable
internal fun FileCardContent(fileName: String) {
    CardRow(minWidth = 168.dp, maxWidth = 240.dp) {
        Text(text = stringResource(R.string.chat_file_card_glyph), fontSize = FileGlyphSize)
        Column(modifier = Modifier.weight(1f, fill = false)) {
            Text(
                text = fileName,
                fontSize = AppType.Subhead,
                fontWeight = FontWeight.Medium,
                color = AppColors.TextPrimary,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            CardHint(stringResource(R.string.chat_file_card_hint))
        }
    }
}
