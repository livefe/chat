package com.nearby.chat.feature.chat.thread

import android.net.Uri
import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.SendStatus
import com.nearby.chat.core.chat.applyIncomingChat
import com.nearby.chat.core.chat.applyRecall
import com.nearby.chat.core.chat.encodeChatFile
import com.nearby.chat.core.chat.encodeChatLocation
import com.nearby.chat.core.chat.markSendFailed
import com.nearby.chat.core.chat.mergeMessages
import com.nearby.chat.core.chat.newOptimisticMessage
import com.nearby.chat.core.chat.replyQuoteText
import com.nearby.chat.core.format.sanitizeNickname
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.ui.TextResolver
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.data.repo.UploadRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

class ChatThread internal constructor(
    private val scope: CoroutineScope,
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val uploadRepository: UploadRepository,
    private val toaster: Toaster,
    private val textResolver: TextResolver,
    private val messageScope: Int,
    private val accepts: (ChatEvent, MessageItem) -> Boolean,
    private val route: () -> ChatRoute?,
    @StringRes private val notReadyMessage: Int?,
) {
    private val _state = MutableStateFlow(ChatThreadState())
    val state: StateFlow<ChatThreadState> = _state.asStateFlow()

    private var loadMoreJob: Job? = null

    init {
        scope.launch {
            _state.update { it.copy(selfId = sessionRepository.selfId()) }
        }
        scope.launch {
            chatRepository.events.collect(::onEvent)
        }
        scope.launch {
            chatRepository.sendFailed.collect { event ->
                val owned = _state.value.messages.any {
                    it.clientMsgId == event.clientMsgId && it.id <= 0 && it.sendStatus == SendStatus.SENDING
                }
                if (!owned) return@collect
                if (event.reason.isBlank()) toaster.fail(R.string.chat_send_failed) else toaster.fail(UiText.of(event.reason))
                _state.update { s -> s.copy(messages = markSendFailed(s.messages, event.clientMsgId)) }
            }
        }
    }

    private fun onEvent(event: ChatEvent) {
        val recalledId = (event as? ChatEvent.Recall)?.item?.id
        _state.update { s ->
            val next = applyIncomingChat(s.messages, event) { item -> accepts(event, item) } ?: return@update s
            s.copy(
                messages = next,
                replyTo = if (recalledId != null) s.replyTo?.takeUnless { it.id == recalledId } else s.replyTo,
            )
        }
    }

    fun onInputChange(value: String) = _state.update { it.copy(input = value) }

    fun setReplyTo(message: MessageItem?) {
        if (message != null && (message.id <= 0 || message.recalled)) return
        _state.update { it.copy(replyTo = message) }
    }

    fun clear() {
        loadMoreJob?.cancel()
        _state.update { it.copy(messages = emptyList(), hasMore = true) }
    }

    fun replaceMessages(messages: List<MessageItem>, hasMore: Boolean) {
        _state.update { it.copy(messages = messages, hasMore = hasMore) }
    }

    fun updateMessages(hasMore: Boolean, transform: (List<MessageItem>) -> List<MessageItem>) {
        _state.update { it.copy(messages = transform(it.messages), hasMore = hasMore) }
    }

    fun mergeLatest(page: List<MessageItem>, limit: Int = ChatLimits.CHAT_PAGE_LIMIT) {
        val sorted = page.sortedBy { it.id }
        updateMessages(hasMore = page.size >= limit) { mergeMessages(it, sorted) }
    }

    fun loadOlder(
        requireHasMore: Boolean = true,
        limit: Int = ChatLimits.CHAT_PAGE_LIMIT,
        filter: (MessageItem) -> Boolean = { true },
        onError: ((Throwable) -> Unit)? = null,
        fetch: suspend (beforeId: Long) -> List<MessageItem>,
    ) {
        val s = _state.value
        if (requireHasMore && !s.hasMore) return
        if (loadMoreJob?.isActive == true) return
        val beforeId = s.messages.filter { it.id > 0 }.minOfOrNull { it.id } ?: return
        loadMoreJob = scope.launch {
            runCatching { fetch(beforeId) }
                .onSuccess { page ->
                    val older = page.filter(filter).sortedBy { it.id }
                    updateMessages(hasMore = page.size >= limit) { mergeMessages(it, older).filter(filter) }
                }
                .onFailure { e -> onError?.invoke(e) }
        }
    }

    fun recall(id: Long) {
        if (id <= 0) return
        scope.launch {
            runCatching { chatRepository.recallMessage(id) }
                .onSuccess {
                    _state.update { s ->
                        s.copy(
                            messages = applyRecall(s.messages, id),
                            replyTo = s.replyTo?.takeUnless { it.id == id },
                        )
                    }
                }
                .onFailure { e -> toaster.fail(e, R.string.chat_recall_failed) }
        }
    }

    fun sendText() {
        val s = _state.value
        val content = s.input.trim()
        if (content.isBlank() || s.uploading) return
        val target = route() ?: return
        dispatch(target, content, MessageContentTypes.TEXT, clearInput = true)
    }

    fun sendImage(uri: Uri) {
        val target = readyRoute() ?: return
        if (_state.value.uploading) return
        upload {
            val media = uploadRepository.uploadChatMedia(uri)
            dispatch(target, media.mediaId.toString(), media.contentType, clearInput = false)
        }
    }

    fun sendFile(uri: Uri) {
        val target = readyRoute() ?: return
        if (_state.value.uploading) return
        upload {
            val media = uploadRepository.uploadChatFile(uri)
            dispatch(target, encodeChatFile(media.mediaId, media.fileName), MessageContentTypes.FILE, clearInput = false)
        }
    }

    fun sendLocation(lat: Double, lng: Double, placeName: String) {
        val target = readyRoute() ?: return
        if (_state.value.uploading) return
        dispatch(target, encodeChatLocation(lat, lng, placeName), MessageContentTypes.LOCATION, clearInput = false)
    }

    suspend fun uploadVoice(uri: Uri, durationMs: Long): String = uploadRepository.uploadChatVoice(uri, durationMs).url

    fun sendVoice(contentJson: String) {
        val target = readyRoute() ?: return
        if (contentJson.isBlank() || _state.value.uploading) return
        dispatch(target, contentJson, MessageContentTypes.VOICE, clearInput = false)
    }

    private fun readyRoute(): ChatRoute? {
        val target = route()
        if (target == null) notReadyMessage?.let { toaster.fail(it) }
        return target
    }

    private fun upload(block: suspend () -> Unit) {
        scope.launch {
            _state.update { it.copy(uploading = true) }
            runCatching { block() }.onFailure { e -> toaster.fail(e, R.string.upload_failed) }
            _state.update { it.copy(uploading = false) }
        }
    }

    private fun dispatch(target: ChatRoute, content: String, contentType: Int, clearInput: Boolean) {
        val reply = _state.value.replyTo
        scope.launch {
            if (!chatRepository.isConnected()) {
                toaster.fail(R.string.chat_not_connected_retrying)
                if (!chatRepository.awaitConnected()) {
                    toaster.fail(R.string.chat_connect_failed_retry)
                    return@launch
                }
            }
            val me = sessionRepository.currentUser()
            val (clientMsgId, optimistic) = newOptimisticMessage(
                selfId = _state.value.selfId,
                selfName = sanitizeNickname(me?.nickname),
                selfAvatar = me?.avatarUrl.orEmpty(),
                content = content,
                contentType = contentType,
                scope = messageScope,
                roomId = target.roomId,
                postId = target.postId,
                toUserId = target.toUserId,
                replyTo = reply,
                replyToContent = reply?.let { textResolver.resolve(replyQuoteText(it.contentType, it.content)) }.orEmpty(),
            )
            _state.update {
                it.copy(
                    input = if (clearInput) "" else it.input,
                    messages = it.messages + optimistic,
                    replyTo = null,
                )
            }
            if (!send(target, content, contentType, clientMsgId, reply?.id ?: 0)) {
                toaster.fail(R.string.chat_send_failed_retry)
                _state.update { st -> st.copy(messages = markSendFailed(st.messages, clientMsgId)) }
            }
        }
    }

    private fun send(target: ChatRoute, content: String, contentType: Int, clientMsgId: String, replyToId: Long): Boolean = when (messageScope) {
        MessageScopes.NEARBY -> chatRepository.sendNearby(
            content,
            target.lat ?: 0.0,
            target.lng ?: 0.0,
            contentType,
            clientMsgId,
            replyToId,
        )

        MessageScopes.ROOM -> chatRepository.sendRoom(
            target.roomId,
            content,
            target.lat,
            target.lng,
            contentType,
            clientMsgId,
            replyToId,
        )

        MessageScopes.PRIVATE -> chatRepository.sendPrivate(target.toUserId, content, contentType, clientMsgId, replyToId)

        MessageScopes.POST -> chatRepository.sendPost(target.postId, content, contentType, clientMsgId, replyToId)

        else -> false
    }
}
