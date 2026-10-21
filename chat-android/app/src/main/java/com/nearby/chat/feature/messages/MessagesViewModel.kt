package com.nearby.chat.feature.messages

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.format.sanitizeNickname
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.paging.OffsetPageTracker
import com.nearby.chat.core.ui.TextResolver
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.local.NotificationPromptStore
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.SessionRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MessagesViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
    private val textResolver: TextResolver,
    private val notificationPromptStore: NotificationPromptStore,
) : ViewModel() {
    private val _uiState = MutableStateFlow(MessagesUiState())
    private val pager = OffsetPageTracker(ChatLimits.CONVERSATION_PAGE)

    private val _notificationPrompt = MutableStateFlow(false)
    val notificationPrompt: StateFlow<Boolean> = _notificationPrompt

    val uiState: StateFlow<MessagesUiState> = combine(
        _uiState,
        chatRepository.unreadByPeer,
    ) { state, unread ->
        state.copy(
            list = state.list.map { item ->
                item.copy(unreadCount = unread[item.peerId] ?: item.unreadCount)
            },
        )
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5_000), MessagesUiState())

    init {
        viewModelScope.launch {
            chatRepository.events.collect { event ->
                when (event) {
                    is ChatEvent.Message -> onPrivateMessage(event.scope, event.item)
                    is ChatEvent.Recall -> onPrivateRecall(event.scope, event.item)
                    else -> Unit
                }
            }
        }
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true) }
            pager.reset()
            runCatching { chatRepository.conversationList(pager.page, pager.pageSize) }
                .onSuccess { (total, list) ->
                    val result = pager.apply(emptyList(), total, list.sanitized())
                    _uiState.update {
                        it.copy(list = sortConversations(result.list), hasMore = result.hasMore)
                    }
                }
                .onFailure { e -> toaster.fail(e, R.string.common_load_failed) }
            _uiState.update { it.copy(loading = false) }
        }
    }

    fun checkNotificationPrompt() {
        viewModelScope.launch {
            if (!notificationPromptStore.prompted()) _notificationPrompt.value = true
        }
    }

    fun onNotificationPromptHandled() {
        _notificationPrompt.value = false
        viewModelScope.launch { notificationPromptStore.markPrompted() }
    }

    fun loadMore() {
        val s = _uiState.value
        if (!s.hasMore || s.loadingMore || s.loading) return
        viewModelScope.launch {
            _uiState.update { it.copy(loadingMore = true) }
            runCatching { chatRepository.conversationList(pager.page, pager.pageSize) }
                .onSuccess { (total, list) ->
                    _uiState.update { cur ->
                        val result = pager.apply(cur.list, total, list.sanitized())
                        val seen = HashSet<Long>()
                        cur.copy(
                            list = sortConversations(result.list.filter { seen.add(it.peerId) }),
                            hasMore = result.hasMore,
                        )
                    }
                }
                .onFailure { e -> toaster.fail(e, R.string.common_load_failed) }
            _uiState.update { it.copy(loadingMore = false) }
        }
    }

    fun setPinned(peerId: Long, pinned: Boolean) {
        if (peerId <= 0L) return
        val prev = _uiState.value.list
        _uiState.update { s ->
            s.copy(list = sortConversations(prev.map { if (it.peerId == peerId) it.copy(pinned = pinned) else it }))
        }
        viewModelScope.launch {
            runCatching { chatRepository.pinConversation(peerId, pinned) }
                .onSuccess {
                    toaster.success(if (pinned) R.string.conversation_pinned else R.string.conversation_unpinned)
                }
                .onFailure { e ->
                    _uiState.update { it.copy(list = prev) }
                    toaster.fail(e, if (pinned) R.string.conversation_pin_failed else R.string.conversation_unpin_failed)
                }
        }
    }

    fun deleteConversation(peerId: Long) {
        if (peerId <= 0L) return
        viewModelScope.launch {
            runCatching { chatRepository.deleteConversation(peerId) }
                .onSuccess {
                    _uiState.update { s -> s.copy(list = s.list.filterNot { c -> c.peerId == peerId }) }
                    toaster.success(R.string.conversation_deleted)
                }
                .onFailure { e -> toaster.fail(e, R.string.common_delete_failed) }
        }
    }

    private suspend fun onPrivateMessage(scope: Int, item: MessageItem) {
        if (scope != MessageScopes.PRIVATE && item.scope != MessageScopes.PRIVATE) return
        val selfId = sessionRepository.selfId()
        if (selfId <= 0L) return
        val peerId = if (item.fromUserId == selfId) item.toUserId else item.fromUserId
        if (peerId <= 0L) return
        if (!chatRepository.messageVisible(peerId, item)) return
        upsertConversationPreview(
            peerId = peerId,
            content = textResolver.resolve(messagePreview(item.contentType, item.content)),
            activeAt = item.createdAt,
            peerName = sanitizeNickname(item.fromName.takeIf { item.fromUserId != selfId }.orEmpty()),
            peerAvatar = item.fromAvatar.takeIf { item.fromUserId != selfId }.orEmpty(),
        )
    }

    private suspend fun onPrivateRecall(scope: Int, item: MessageItem) {
        if (scope != MessageScopes.PRIVATE && item.scope != MessageScopes.PRIVATE) return
        val selfId = sessionRepository.selfId()
        if (selfId <= 0L) return
        val peerId = if (item.fromUserId == selfId) item.toUserId else item.fromUserId
        if (peerId <= 0L) return
        patchConversationPreviewOnRecall(peerId, item)
    }

    private fun upsertConversationPreview(
        peerId: Long,
        content: String,
        activeAt: Long,
        peerName: String,
        peerAvatar: String,
    ) {
        _uiState.update { s ->
            val existing = s.list.firstOrNull { it.peerId == peerId }
            val updated = ConversationItem(
                peerId = peerId,
                peerNickname = peerName.ifBlank { sanitizeNickname(existing?.peerNickname) },
                peerAvatar = peerAvatar.ifBlank { existing?.peerAvatar.orEmpty() },
                lastContent = content,
                lastActiveAt = activeAt.takeIf { it > 0 } ?: (existing?.lastActiveAt ?: 0L),
                unreadCount = existing?.unreadCount ?: 0,
                pinned = existing?.pinned == true,
            )
            s.copy(list = sortConversations(listOf(updated) + s.list.filterNot { it.peerId == peerId }))
        }
    }

    private fun patchConversationPreviewOnRecall(peerId: Long, item: MessageItem) {
        val recalledPreview = textResolver.resolve(messagePreview(item.contentType, item.content))
        val recalledText = textResolver.resolve(UiText.of(R.string.msg_recalled_generic))
        _uiState.update { s ->
            s.copy(
                list = s.list.map { c ->
                    if (c.peerId != peerId) {
                        c
                    } else {
                        val last = c.lastContent
                        val wasLast = last.isBlank() ||
                            last == recalledPreview ||
                            recalledPreview.startsWith(last) ||
                            last.startsWith(recalledPreview)
                        if (wasLast) c.copy(lastContent = recalledText) else c
                    }
                },
            )
        }
    }

    private fun List<ConversationItem>.sanitized(): List<ConversationItem> = map { c -> c.copy(peerNickname = sanitizeNickname(c.peerNickname)) }

    private fun sortConversations(list: List<ConversationItem>): List<ConversationItem> = list.sortedWith(
        compareByDescending<ConversationItem> { it.pinned }
            .thenByDescending { it.lastActiveAt },
    )
}
