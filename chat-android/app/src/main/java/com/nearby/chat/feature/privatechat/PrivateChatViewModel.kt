package com.nearby.chat.feature.privatechat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.mergeMessages
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.feature.chat.thread.ChatRoute
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadFactory
import com.nearby.chat.feature.chat.thread.ChatThreadState
import com.nearby.chat.feature.chat.thread.messageScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PrivateChatViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
    threadFactory: ChatThreadFactory,
) : ViewModel() {
    private val _uiState = MutableStateFlow(PrivateChatUiState())
    val uiState: StateFlow<PrivateChatUiState> = _uiState.asStateFlow()

    private var selfId: Long = 0

    val thread: ChatThread = threadFactory.create(
        scope = viewModelScope,
        messageScope = MessageScopes.PRIVATE,
        accepts = ::acceptsEvent,
        route = {
            val peerId = _uiState.value.peerId
            if (peerId > 0) ChatRoute(toUserId = peerId) else null
        },
    )
    val threadState: StateFlow<ChatThreadState> = thread.state

    init {
        viewModelScope.launch {
            chatRepository.reconnectTick.collect {
                val peerId = _uiState.value.peerId
                if (peerId > 0) loadInitial(peerId)
            }
        }
    }

    fun load(peerId: Long, peerName: String) {
        viewModelScope.launch {
            chatRepository.setActivePrivatePeer(peerId)
            selfId = sessionRepository.selfId()
            if (_uiState.value.peerId != peerId) thread.clear()
            _uiState.update { it.copy(peerId = peerId, peerName = peerName) }
            loadInitial(peerId)
            runCatching { chatRepository.markRead(peerId) }
        }
    }

    fun loadMore() {
        val peerId = _uiState.value.peerId
        thread.loadOlder(
            requireHasMore = false,
            filter = { chatRepository.messageVisible(peerId, it) },
            onError = { e -> toaster.fail(e, R.string.common_load_failed) },
        ) { beforeId ->
            chatRepository.privateMessages(peerId, beforeId, ChatLimits.CHAT_PAGE_LIMIT)
        }
    }

    override fun onCleared() {
        val peerId = _uiState.value.peerId
        chatRepository.clearActivePrivatePeer(peerId)
        chatRepository.markReadAsync(peerId)
        super.onCleared()
    }

    private fun acceptsEvent(event: ChatEvent, item: MessageItem): Boolean {
        val peerId = _uiState.value.peerId
        if (peerId == 0L) return false
        val isPrivateScope = event.messageScope == MessageScopes.PRIVATE || item.scope == MessageScopes.PRIVATE
        val isThisPeer = (item.fromUserId == peerId && item.toUserId == selfId) ||
            (item.fromUserId == selfId && item.toUserId == peerId)
        if (!isPrivateScope || !isThisPeer) return false
        return event is ChatEvent.Recall || chatRepository.messageVisible(peerId, item)
    }

    private suspend fun loadInitial(peerId: Long) {
        runCatching {
            chatRepository.privateMessages(peerId, 0, ChatLimits.CHAT_PAGE_LIMIT)
        }.onSuccess { page ->
            val visible: (MessageItem) -> Boolean = { chatRepository.messageVisible(peerId, it) }
            val list = page.filter(visible).sortedBy { it.id }
            thread.updateMessages(hasMore = page.size >= ChatLimits.CHAT_PAGE_LIMIT) { current ->
                mergeMessages(current, list).filter(visible)
            }
        }.onFailure { e -> toaster.fail(e, R.string.common_load_failed) }
    }
}
