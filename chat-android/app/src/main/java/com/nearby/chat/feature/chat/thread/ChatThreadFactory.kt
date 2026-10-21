package com.nearby.chat.feature.chat.thread

import androidx.annotation.StringRes
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.ui.TextResolver
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.data.repo.UploadRepository
import kotlinx.coroutines.CoroutineScope
import javax.inject.Inject

class ChatThreadFactory @Inject constructor(
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val uploadRepository: UploadRepository,
    private val toaster: Toaster,
    private val textResolver: TextResolver,
) {
    fun create(
        scope: CoroutineScope,
        messageScope: Int,
        accepts: (ChatEvent, MessageItem) -> Boolean,
        route: () -> ChatRoute?,
        @StringRes notReadyMessage: Int? = null,
    ): ChatThread = ChatThread(
        scope = scope,
        chatRepository = chatRepository,
        sessionRepository = sessionRepository,
        uploadRepository = uploadRepository,
        toaster = toaster,
        textResolver = textResolver,
        messageScope = messageScope,
        accepts = accepts,
        route = route,
        notReadyMessage = notReadyMessage,
    )
}
