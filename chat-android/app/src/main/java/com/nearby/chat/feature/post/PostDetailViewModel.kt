package com.nearby.chat.feature.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.core.ui.UiText
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.PostRepository
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
class PostDetailViewModel @Inject constructor(
    private val postRepository: PostRepository,
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
    threadFactory: ChatThreadFactory,
) : ViewModel() {
    private val _uiState = MutableStateFlow(PostDetailUiState())
    val uiState: StateFlow<PostDetailUiState> = _uiState.asStateFlow()

    private var postId: Long = 0

    val thread: ChatThread = threadFactory.create(
        scope = viewModelScope,
        messageScope = MessageScopes.POST,
        accepts = { event, item ->
            (event.messageScope == MessageScopes.POST || item.scope == MessageScopes.POST) && item.postId == postId
        },
        route = { if (postId > 0) ChatRoute(postId = postId) else null },
    )
    val threadState: StateFlow<ChatThreadState> = thread.state

    init {
        viewModelScope.launch {
            chatRepository.events.collect { event ->
                if (event is ChatEvent.PostPresence && event.postId == postId) {
                    _uiState.update { it.copy(onlineCount = event.onlineCount) }
                }
            }
        }
        viewModelScope.launch {
            chatRepository.reconnectTick.collect {
                if (postId > 0 && _uiState.value.showChat) refreshMessages()
            }
        }
    }

    fun load(id: Long) {
        if (postId != id) {
            if (postId > 0) chatRepository.unsubscribePost(postId)
            thread.clear()
            _uiState.update {
                it.copy(discussionLoaded = false, showChat = false, onlineCount = 0)
            }
        }
        postId = id
        viewModelScope.launch {
            _uiState.update { it.copy(selfId = sessionRepository.selfId()) }
            runCatching { postRepository.detail(id) }
                .onSuccess { detail -> _uiState.update { it.copy(detail = detail) } }
                .onFailure { e -> toaster.fail(e, R.string.common_load_failed) }
        }
    }

    fun signup() {
        viewModelScope.launch {
            runCatching { postRepository.signup(postId) }
                .onSuccess { status ->
                    toaster.success(
                        if (status == PostSignupRules.STATUS_WAITLISTED) {
                            R.string.signup_waitlisted_toast
                        } else {
                            R.string.signup_success
                        },
                    )
                    load(postId)
                }
                .onFailure { e -> toaster.fail(e, R.string.signup_failed) }
        }
    }

    fun cancelSignup() {
        val wasWaitlisted = _uiState.value.detail?.mySignupStatus == PostSignupRules.STATUS_WAITLISTED
        viewModelScope.launch {
            runCatching { postRepository.signupCancel(postId) }
                .onSuccess {
                    toaster.success(if (wasWaitlisted) R.string.signup_waitlist_cancelled else R.string.signup_cancelled)
                    load(postId)
                }
                .onFailure { e -> toaster.fail(e, R.string.signup_cancel_failed) }
        }
    }

    fun openChat() {
        if (!_uiState.value.showChat) toggleChat()
    }

    fun closeChat() {
        if (_uiState.value.showChat) toggleChat()
    }

    private fun toggleChat() {
        val show = !_uiState.value.showChat
        if (!show) {
            chatRepository.unsubscribePost(postId)
            _uiState.update { it.copy(showChat = false, onlineCount = 0) }
            return
        }
        _uiState.update { it.copy(showChat = true) }
        viewModelScope.launch {
            val joined = chatRepository.awaitSubscribePost(postId)
            if (joined.isFailure) {
                val reason = joined.exceptionOrNull()?.message
                if (reason == null) toaster.fail(R.string.chat_join_post_failed) else toaster.fail(UiText.of(reason))
                _uiState.update { it.copy(showChat = false, onlineCount = 0) }
                return@launch
            }
            if (!refreshMessages()) {
                chatRepository.unsubscribePost(postId)
                _uiState.update { it.copy(showChat = false, onlineCount = 0) }
            }
        }
    }

    fun loadMoreMessages() {
        if (postId <= 0) return
        thread.loadOlder { beforeId ->
            chatRepository.postMessages(postId, beforeId, ChatLimits.CHAT_PAGE_LIMIT)
        }
    }

    private suspend fun refreshMessages(): Boolean {
        if (postId <= 0) return false
        return runCatching {
            chatRepository.postMessages(postId, 0, ChatLimits.CHAT_PAGE_LIMIT)
        }.onSuccess { page ->
            thread.mergeLatest(page)
            _uiState.update { it.copy(discussionLoaded = true) }
        }.onFailure { e ->
            toaster.fail(e, R.string.chat_open_discussion_failed)
        }.isSuccess
    }

    override fun onCleared() {
        if (postId > 0) chatRepository.unsubscribePost(postId)
        super.onCleared()
    }
}
