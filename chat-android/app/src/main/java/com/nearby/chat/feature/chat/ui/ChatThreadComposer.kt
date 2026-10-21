package com.nearby.chat.feature.chat.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadState

@Composable
fun ChatThreadComposer(
    thread: ChatThread,
    state: ChatThreadState,
    onPickLocation: () -> Unit,
    sendEnabled: Boolean,
    moreOpen: Boolean,
    onMoreOpenChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
    canCompose: Boolean = true,
) {
    ChatComposerBar(
        input = state.input,
        onInputChange = thread::onInputChange,
        onSend = thread::sendText,
        onImagePicked = thread::sendImage,
        onFilePicked = thread::sendFile,
        onPickLocation = onPickLocation,
        onVoiceRecorded = thread::sendVoice,
        uploadVoice = thread::uploadVoice,
        sendEnabled = sendEnabled,
        canCompose = canCompose,
        uploading = state.uploading,
        replyTo = state.replyTo,
        onClearReply = { thread.setReplyTo(null) },
        moreOpen = moreOpen,
        onMoreOpenChange = onMoreOpenChange,
        modifier = modifier,
    )
}
