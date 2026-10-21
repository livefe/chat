package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Popup
import androidx.compose.ui.window.PopupProperties
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.SendStatus
import com.nearby.chat.core.chat.chatFileDisplayName
import com.nearby.chat.core.chat.chatMediaDisplayUrl
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.chat.openChatLocationMap
import com.nearby.chat.core.chat.parseChatLocation
import com.nearby.chat.core.chat.parseChatPost
import com.nearby.chat.core.chat.parseChatVoice
import com.nearby.chat.core.chat.truncateFileDisplayName
import com.nearby.chat.core.media.ChatFileDownload
import com.nearby.chat.core.media.MediaUrl
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.AppConfirmDialog
import com.nearby.chat.core.ui.LocalToaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.core.ui.UserAvatar
import com.nearby.chat.core.ui.asString

private val MessageAvatarSize = 40.dp
private val MessageAvatarGap = AppSpacing.S
private val MessageMaxWidth = 280.dp
private val MenuOffsetY = (-44).dp
private const val REPORT_PREVIEW_MAX = 40

@Composable
fun MessageBubble(
    message: MessageItem,
    isMine: Boolean,
    onAvatarClick: ((Long) -> Unit)? = null,
    hideName: Boolean = false,
    onReply: ((MessageItem) -> Unit)? = null,
    onForward: ((MessageItem) -> Unit)? = null,
    onReport: ((MessageItem) -> Unit)? = null,
    onRecall: ((MessageItem) -> Unit)? = null,
    onPostClick: ((Long) -> Unit)? = null,
) {
    if (message.scope == MessageScopes.SYSTEM_HINT) {
        CenterHint(message.content)
        return
    }
    if (message.recalled) {
        CenterHint(
            when {
                isMine -> stringResource(R.string.msg_recalled_self)
                hideName -> stringResource(R.string.msg_recalled_peer)
                message.fromName.isNotBlank() -> stringResource(R.string.msg_recalled_named, message.fromName)
                else -> stringResource(R.string.msg_recalled_generic)
            },
        )
        return
    }
    val context = LocalContext.current
    val density = LocalDensity.current
    val toaster = LocalToaster.current
    val showName = !hideName && !isMine && message.fromName.isNotBlank()
    val contentType = message.contentType
    val isFile = contentType == MessageContentTypes.FILE
    val location = if (contentType == MessageContentTypes.LOCATION) parseChatLocation(message.content) else null
    val post = if (contentType == MessageContentTypes.POST) parseChatPost(message.content) else null
    val sending = message.sendStatus == SendStatus.SENDING
    val failed = message.sendStatus == SendStatus.FAILED
    var previewUrl by remember { mutableStateOf<String?>(null) }
    var previewVideoUrl by remember { mutableStateOf<String?>(null) }
    var menuOpen by remember { mutableStateOf(false) }
    var reportConfirm by remember { mutableStateOf(false) }
    var fileDownloadConfirm by remember { mutableStateOf(false) }
    val canRecall = isMine &&
        onRecall != null &&
        message.id > 0 &&
        (System.currentTimeMillis() / 1000 - message.createdAt) <= ChatLimits.RECALL_WINDOW_SEC
    val canOpenMenu = onReply != null && message.id > 0
    val openMenu: (() -> Unit)? = if (canOpenMenu) ({ menuOpen = true }) else null
    val avatarInset = MessageAvatarSize + MessageAvatarGap

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = AppSpacing.M, vertical = AppSpacing.Xs)
            .then(if (sending) Modifier.alpha(SENDING_ALPHA) else Modifier),
        horizontalAlignment = if (isMine) Alignment.End else Alignment.Start,
    ) {
        if (message.replyToId > 0) {
            ReplyQuoteBlock(
                name = message.replyToFromName,
                content = message.replyToContent,
                contentType = message.replyToContentType,
                onImageClick = { previewUrl = it },
                onVideoClick = { previewVideoUrl = it },
                modifier = Modifier
                    .widthIn(max = MessageMaxWidth)
                    .padding(
                        start = if (isMine) 0.dp else avatarInset,
                        end = if (isMine) avatarInset else 0.dp,
                    ),
            )
            Spacer(modifier = Modifier.size(AppSpacing.Xs))
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = if (isMine) Arrangement.End else Arrangement.Start,
            verticalAlignment = Alignment.Top,
        ) {
            if (!isMine) {
                UserAvatar(
                    url = message.fromAvatar,
                    size = MessageAvatarSize,
                    onClick = onAvatarClick?.let { cb -> { cb(message.fromUserId) } },
                )
                Spacer(modifier = Modifier.size(MessageAvatarGap))
            }
            Box {
                Column(
                    horizontalAlignment = if (isMine) Alignment.End else Alignment.Start,
                    modifier = Modifier.widthIn(max = MessageMaxWidth),
                ) {
                    if (showName) {
                        SenderName(message.fromName)
                        Spacer(modifier = Modifier.size(AppSpacing.Xs))
                    }
                    when (contentType) {
                        MessageContentTypes.IMAGE -> ImageMessage(
                            url = chatMediaDisplayUrl(message.content),
                            failed = failed,
                            onPreview = { previewUrl = it },
                            onLongClick = openMenu,
                        )

                        MessageContentTypes.VOICE -> {
                            val voice = parseChatVoice(message.content)
                            if (voice != null) {
                                val keySuffix = message.id.takeIf { it > 0 }?.toString()
                                    ?: message.clientMsgId.ifBlank { voice.url }
                                VoiceMessageBubble(
                                    voiceUrl = MediaUrl.resolve(voice.url),
                                    durationMs = voice.durationMs,
                                    isMine = isMine,
                                    playKey = "msg-$keySuffix",
                                    sending = sending,
                                    failed = failed,
                                    onLongClick = openMenu,
                                )
                            } else {
                                Text(
                                    text = stringResource(R.string.msg_preview_voice),
                                    color = AppColors.TextMuted,
                                    fontSize = AppType.Secondary,
                                    modifier = Modifier.padding(horizontal = AppSpacing.Xs, vertical = 6.dp),
                                )
                            }
                        }

                        MessageContentTypes.VIDEO -> ChatVideoBubble(
                            url = chatMediaDisplayUrl(message.content),
                            sending = sending,
                            failed = failed,
                            onLongClick = openMenu,
                        )

                        else -> MessageContentBubble(
                            content = message.content,
                            isMine = isMine,
                            failed = failed,
                            location = location,
                            post = post,
                            isFile = isFile,
                            onClick = when {
                                location != null -> ({ openChatLocationMap(context, location) })
                                post != null -> ({ onPostClick?.invoke(post.postId) })
                                isFile -> ({ fileDownloadConfirm = true })
                                else -> null
                            },
                            onLongClick = openMenu,
                        )
                    }
                }
                if (menuOpen) {
                    Popup(
                        alignment = if (isMine) Alignment.TopEnd else Alignment.TopStart,
                        offset = IntOffset(x = 0, y = with(density) { MenuOffsetY.roundToPx() }),
                        onDismissRequest = { menuOpen = false },
                        properties = PopupProperties(
                            focusable = true,
                            dismissOnBackPress = true,
                            dismissOnClickOutside = true,
                        ),
                    ) {
                        MessageActionMenu(
                            onCopy = {
                                copyMessage(context, toaster, message)
                                menuOpen = false
                            },
                            onForward = onForward?.let { handler ->
                                {
                                    menuOpen = false
                                    handler(message)
                                }
                            },
                            onQuote = {
                                onReply?.invoke(message)
                                menuOpen = false
                            },
                            onRecall = if (canRecall) {
                                {
                                    menuOpen = false
                                    onRecall?.invoke(message)
                                }
                            } else {
                                null
                            },
                            onReport = if (!isMine && onReport != null) {
                                {
                                    menuOpen = false
                                    reportConfirm = true
                                }
                            } else {
                                null
                            },
                        )
                    }
                }
            }
            if (isMine) {
                Spacer(modifier = Modifier.size(MessageAvatarGap))
                UserAvatar(
                    url = message.fromAvatar,
                    size = MessageAvatarSize,
                    onClick = onAvatarClick?.let { cb -> { cb(message.fromUserId) } },
                )
            }
        }
    }

    if (reportConfirm) {
        val preview = UiText.Ellipsized(messagePreview(contentType, message.content), REPORT_PREVIEW_MAX).asString()
        AppConfirmDialog(
            title = stringResource(R.string.report_message_title),
            message = stringResource(R.string.report_message_confirm, preview),
            danger = true,
            onConfirm = {
                reportConfirm = false
                onReport?.invoke(message)
            },
            onDismiss = { reportConfirm = false },
        )
    }

    if (fileDownloadConfirm) {
        val name = truncateFileDisplayName(
            chatFileDisplayName(message.content).ifBlank { stringResource(R.string.chat_file_default_name) },
        )
        AppConfirmDialog(
            title = stringResource(R.string.download_file_title),
            message = stringResource(R.string.download_file_confirm, name),
            confirmText = stringResource(R.string.download_action),
            onConfirm = {
                fileDownloadConfirm = false
                ChatFileDownload.start(context, toaster, message.content)
            },
            onDismiss = { fileDownloadConfirm = false },
        )
    }

    previewUrl?.let { url ->
        ImagePreviewDialog(url = url, onDismiss = { previewUrl = null })
    }
    previewVideoUrl?.let { url ->
        ChatVideoPlayer(url = url, onDismiss = { previewVideoUrl = null })
    }
}
