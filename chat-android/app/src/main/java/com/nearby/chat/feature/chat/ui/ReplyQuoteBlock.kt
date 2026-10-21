package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.displayName
import com.nearby.chat.core.chat.openChatLocationMap
import com.nearby.chat.core.chat.parseChatLocation
import com.nearby.chat.core.chat.parseChatVoice
import com.nearby.chat.core.chat.replyQuoteMediaUrl
import com.nearby.chat.core.chat.voiceQuoteLabel
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.asString

private val ReplyQuoteBg = AppColors.NeutralContainer
private val ReplyQuoteShape = RoundedCornerShape(4.dp)
private val ReplyQuoteVPad = 5.dp
private val ReplyQuoteLineHeight = 15.sp

@Composable
internal fun ReplyQuoteBlock(
    name: String,
    content: String,
    contentType: Int,
    modifier: Modifier = Modifier,
    onImageClick: ((String) -> Unit)? = null,
    onVideoClick: ((String) -> Unit)? = null,
) {
    val context = LocalContext.current
    val label = name.ifBlank { stringResource(R.string.message_default_sender) }
    val location = if (contentType == MessageContentTypes.LOCATION) parseChatLocation(content) else null
    val voice = if (contentType == MessageContentTypes.VOICE) parseChatVoice(content) else null
    val mediaUrl = replyQuoteMediaUrl(contentType, content)
    val isVideo = contentType == MessageContentTypes.VIDEO && mediaUrl.isNotBlank()
    val body = when {
        location != null -> location.displayName()

        contentType == MessageContentTypes.VOICE -> when {
            voice != null -> voiceQuoteLabel(voice.durationMs).asString()
            content.isNotBlank() -> content
            else -> stringResource(R.string.msg_preview_voice)
        }

        content.isNotBlank() -> content

        else -> stringResource(R.string.chat_reply_deleted)
    }
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(AppSpacing.Xs),
        modifier = modifier
            .clip(ReplyQuoteShape)
            .background(ReplyQuoteBg)
            .padding(horizontal = AppSpacing.S, vertical = ReplyQuoteVPad),
    ) {
        Text(
            text = stringResource(R.string.chat_reply_prefix, label),
            fontSize = AppType.Micro,
            lineHeight = ReplyQuoteLineHeight,
            color = AppColors.TextMuted,
            maxLines = 1,
        )
        if (mediaUrl.isNotBlank()) {
            if (isVideo) {
                ReplyVideoThumb(
                    url = mediaUrl,
                    onClick = onVideoClick?.let { cb -> { cb(mediaUrl) } },
                )
            } else {
                AsyncImage(
                    model = mediaUrl,
                    contentDescription = stringResource(R.string.upload_kind_image),
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .size(ReplyThumbSize)
                        .clip(ReplyThumbShape)
                        .background(AppColors.PageBg)
                        .then(
                            if (onImageClick != null) {
                                Modifier.clickable { onImageClick(mediaUrl) }
                            } else {
                                Modifier
                            },
                        ),
                )
            }
        } else {
            Text(
                text = body,
                fontSize = AppType.Micro,
                lineHeight = ReplyQuoteLineHeight,
                color = AppColors.TextMuted,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier
                    .weight(1f, fill = false)
                    .then(
                        if (location != null) {
                            Modifier.clickable { openChatLocationMap(context, location) }
                        } else {
                            Modifier
                        },
                    ),
            )
        }
    }
}
