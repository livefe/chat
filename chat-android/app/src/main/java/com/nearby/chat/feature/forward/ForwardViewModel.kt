package com.nearby.chat.feature.forward

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.format.sanitizeNickname
import com.nearby.chat.core.share.PostShare
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID
import javax.inject.Inject

@HiltViewModel
class ForwardViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val locationRepository: LocationRepository,
    private val sessionHolder: ForwardSessionHolder,
    private val postShare: PostShare,
    private val toaster: Toaster,
) : ViewModel() {
    val session: ForwardSession? = sessionHolder.peek()?.takeIf { it.content.isNotBlank() }

    private val _uiState = MutableStateFlow(ForwardUiState())
    val uiState: StateFlow<ForwardUiState> = _uiState.asStateFlow()

    private val _done = Channel<Unit>(Channel.CONFLATED)
    val done: Flow<Unit> = _done.receiveAsFlow()

    private var chatsPage = 1

    init {
        if (session != null) {
            viewModelScope.launch { ensureLocation() }
            openChats()
        }
    }

    fun onQuery(value: String) = _uiState.update { it.copy(query = value) }

    fun selectTab(tab: ForwardTab) {
        if (_uiState.value.tab == tab) return
        _uiState.update { it.copy(query = "") }
        when (tab) {
            ForwardTab.Rooms -> openRooms()
            ForwardTab.Chats -> openChats()
            ForwardTab.Dest -> _uiState.update { it.copy(tab = ForwardTab.Dest) }
        }
    }

    fun loadMoreChats() {
        if (_uiState.value.chatsLoadingMore || !_uiState.value.chatsHasMore) return
        loadChats(reset = false)
    }

    fun sendNearby() {
        val s = session ?: return
        viewModelScope.launch {
            if (!ensureLocation()) {
                toaster.fail(R.string.forward_choose_location)
                return@launch
            }
            val state = _uiState.value
            dispatch(s.content) { body, clientMsgId ->
                chatRepository.sendNearby(body, state.lat, state.lng, s.contentType, clientMsgId)
            }
        }
    }

    fun sendRoom(roomId: Long) {
        val s = session ?: return
        if (roomId <= 0L) return
        viewModelScope.launch {
            ensureLocation()
            val state = _uiState.value
            dispatch(s.content) { body, clientMsgId ->
                chatRepository.sendRoom(
                    roomId = roomId,
                    content = body,
                    lat = state.lat.takeIf { state.locationReady },
                    lng = state.lng.takeIf { state.locationReady },
                    contentType = s.contentType,
                    clientMsgId = clientMsgId,
                )
            }
        }
    }

    fun sendPrivate(peerId: Long) {
        val s = session ?: return
        if (peerId <= 0L) return
        viewModelScope.launch {
            dispatch(s.content) { body, clientMsgId ->
                chatRepository.sendPrivate(peerId, body, s.contentType, clientMsgId)
            }
        }
    }

    fun saveShareImage() {
        val s = session ?: return
        runSharing {
            postShare.saveImage(
                postId = s.postId,
                title = s.postTitle,
                content = s.shareContent,
                imageUrl = s.cover.takeIf { it.isNotBlank() },
            )
        }
    }

    fun shareExternal() {
        val s = session ?: return
        runSharing {
            postShare.share(
                postId = s.postId,
                title = s.postTitle,
                content = s.shareContent,
                imageUrl = s.cover.takeIf { it.isNotBlank() },
            )
        }
    }

    fun copyLink() {
        val s = session ?: return
        postShare.copyLink(s.postId)
    }

    override fun onCleared() {
        session?.let { sessionHolder.clear(it) }
    }

    private fun runSharing(block: suspend () -> Unit) {
        if (_uiState.value.sharing) return
        _uiState.update { it.copy(sharing = true) }
        viewModelScope.launch {
            try {
                block()
            } finally {
                _uiState.update { it.copy(sharing = false) }
            }
        }
    }

    private fun openRooms() {
        _uiState.update { it.copy(tab = ForwardTab.Rooms) }
        viewModelScope.launch {
            ensureLocation()
            loadRooms()
        }
    }

    private fun openChats() {
        _uiState.update { it.copy(tab = ForwardTab.Chats) }
        loadChats(reset = true)
    }

    private suspend fun ensureLocation(): Boolean {
        if (_uiState.value.locationReady) return true
        val saved = locationRepository.savedLocation()
        val session = chatRepository.lastLocationOrNull()
        val lat = saved?.lat ?: session?.first ?: 0.0
        val lng = saved?.lng ?: session?.second ?: 0.0
        val ready = lat != 0.0 || lng != 0.0
        _uiState.update { it.copy(lat = lat, lng = lng, locationReady = ready) }
        return ready
    }

    private fun loadRooms() {
        val s = _uiState.value
        if (!s.locationReady || s.roomsLoading) return
        viewModelScope.launch {
            _uiState.update { it.copy(roomsLoading = true) }
            runCatching {
                chatRepository.syncLocationForNearby(
                    lat = s.lat,
                    lng = s.lng,
                    radius = ChatLimits.NEARBY_RADIUS,
                    limit = ChatLimits.CHAT_PAGE_LIMIT,
                    withMessages = false,
                )
            }.onSuccess { data ->
                _uiState.update {
                    it.copy(
                        rooms = data.rooms.sortedBy { room -> room.distance },
                        roomsLoading = false,
                    )
                }
            }.onFailure { e ->
                _uiState.update { it.copy(roomsLoading = false) }
                toaster.fail(e, R.string.chat_room_load_failed)
            }
        }
    }

    private fun loadChats(reset: Boolean) {
        if (reset) chatsPage = 1
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    chatsLoading = reset && it.chats.isEmpty(),
                    chatsLoadingMore = !reset,
                )
            }
            runCatching { chatRepository.conversationList(chatsPage, ChatLimits.CONVERSATION_PAGE) }
                .onSuccess { (total, list) ->
                    val items = list.map { c -> c.copy(peerNickname = sanitizeNickname(c.peerNickname)) }
                    _uiState.update { s ->
                        val next = if (reset) items else s.chats + items
                        s.copy(
                            chats = next,
                            chatsHasMore = next.size < total,
                            chatsLoading = false,
                            chatsLoadingMore = false,
                        )
                    }
                    if (items.isNotEmpty()) chatsPage += 1
                }
                .onFailure { e ->
                    _uiState.update { it.copy(chatsLoading = false, chatsLoadingMore = false) }
                    toaster.fail(e, R.string.common_load_failed)
                }
        }
    }

    private suspend fun dispatch(
        content: String,
        send: (content: String, clientMsgId: String) -> Boolean,
    ) {
        val body = content.trim()
        if (body.isEmpty() || _uiState.value.sending) return
        _uiState.update { it.copy(sending = true) }
        chatRepository.connect()
        if (!chatRepository.isConnected() && !chatRepository.awaitConnected()) {
            toaster.fail(R.string.chat_connect_failed_retry)
            _uiState.update { it.copy(sending = false) }
            return
        }
        val clientMsgId = UUID.randomUUID().toString()
        val ok = send(body, clientMsgId)
        _uiState.update { it.copy(sending = false) }
        if (!ok) {
            toaster.fail(R.string.chat_send_failed)
            return
        }
        toaster.success(R.string.forward_done)
        _done.send(Unit)
    }
}
