package com.nearby.chat.feature.chat

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.ChatLimits
import com.nearby.chat.core.chat.MessageScopes
import com.nearby.chat.core.chat.mergeMessages
import com.nearby.chat.core.model.ChatEvent
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.LocationRepository
import com.nearby.chat.feature.chat.thread.ChatRoute
import com.nearby.chat.feature.chat.thread.ChatThread
import com.nearby.chat.feature.chat.thread.ChatThreadFactory
import com.nearby.chat.feature.chat.thread.ChatThreadState
import com.nearby.chat.feature.chat.thread.messageScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

private data class PendingShare(val lat: Double, val lng: Double, val placeName: String)

@HiltViewModel
class RoomDetailViewModel @Inject constructor(
    private val chatRepository: ChatRepository,
    private val locationRepository: LocationRepository,
    private val toaster: Toaster,
    threadFactory: ChatThreadFactory,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val _uiState = MutableStateFlow(RoomDetailUiState())
    val uiState: StateFlow<RoomDetailUiState> = _uiState.asStateFlow()

    private var roomId: Long = savedStateHandle.get<Long>("roomId") ?: 0
    private var loadJob: Job? = null
    private var pendingShare: PendingShare? = null

    val thread: ChatThread = threadFactory.create(
        scope = viewModelScope,
        messageScope = MessageScopes.ROOM,
        accepts = { event, item ->
            (event.messageScope == MessageScopes.ROOM || item.scope == MessageScopes.ROOM) && item.roomId == roomId
        },
        route = {
            if (roomId <= 0) {
                null
            } else {
                val loc = chatRepository.lastLocationOrNull()
                ChatRoute(roomId = roomId, lat = loc?.first, lng = loc?.second)
            }
        },
        notReadyMessage = R.string.chat_room_not_ready,
    )
    val threadState: StateFlow<ChatThreadState> = thread.state

    init {
        viewModelScope.launch {
            chatRepository.events.collect { event ->
                if (event is ChatEvent.RoomPresence && event.roomId == roomId) {
                    _uiState.update { it.copy(onlineCount = event.onlineCount) }
                }
            }
        }
        viewModelScope.launch {
            chatRepository.reconnectTick.collect {
                if (roomId > 0) refreshMessages()
            }
        }
    }

    fun load(id: Long) {
        if (id <= 0) return

        if (id == roomId && _uiState.value.room != null) {
            if (!chatRepository.isRoomSubscribed(id)) {
                viewModelScope.launch {
                    val loc = chatRepository.lastLocationOrNull()
                    chatRepository.awaitSubscribeRoom(id, loc?.first, loc?.second)
                    flushPendingShare()
                }
                return
            }
            flushPendingShare()
            return
        }
        roomId = id
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            var loc = chatRepository.lastLocationOrNull()
            if (loc == null) {
                val saved = runCatching { locationRepository.savedLocation() }.getOrNull()
                if (saved != null) {
                    chatRepository.rememberLocation(saved.lat, saved.lng)
                    loc = saved.lat to saved.lng
                }
            }
            val joined = chatRepository.awaitSubscribeRoom(id, loc?.first, loc?.second)
            if (joined.isFailure) {
                toaster.fail(joined.exceptionOrNull(), R.string.chat_join_room_failed)
                return@launch
            }
            runCatching {
                chatRepository.roomBootstrap(
                    roomId = id,
                    limit = ChatLimits.CHAT_PAGE_LIMIT,
                    withMessages = true,
                    lat = loc?.first,
                    lng = loc?.second,
                )
            }.onSuccess { boot ->
                val sorted = boot.messages.sortedBy { m -> m.id }
                _uiState.update { it.copy(room = boot.room, onlineCount = boot.onlineCount) }
                thread.updateMessages(hasMore = sorted.size >= ChatLimits.CHAT_PAGE_LIMIT) { mergeMessages(it, sorted) }
            }.onFailure { e -> toaster.fail(e, R.string.chat_room_load_failed) }
            flushPendingShare()
        }
    }

    fun enqueueShareLocation(lat: Double, lng: Double, placeName: String) {
        if (roomId > 0 && _uiState.value.room != null) {
            thread.sendLocation(lat, lng, placeName)
            return
        }
        pendingShare = PendingShare(lat, lng, placeName)
    }

    private fun flushPendingShare() {
        val pending = pendingShare ?: return
        pendingShare = null
        thread.sendLocation(pending.lat, pending.lng, pending.placeName)
    }

    private suspend fun refreshMessages() {
        if (roomId <= 0) return
        val loc = chatRepository.lastLocationOrNull()
        runCatching {
            chatRepository.roomMessages(
                roomId = roomId,
                beforeId = 0,
                limit = ChatLimits.CHAT_PAGE_LIMIT,
                lat = loc?.first,
                lng = loc?.second,
            )
        }.onSuccess { list -> thread.mergeLatest(list) }
    }

    fun loadMoreMessages() {
        if (roomId <= 0) return
        val id = roomId
        val loc = chatRepository.lastLocationOrNull()
        thread.loadOlder { beforeId ->
            chatRepository.roomMessages(id, beforeId, ChatLimits.CHAT_PAGE_LIMIT, loc?.first, loc?.second)
        }
    }

    override fun onCleared() {
        chatRepository.unsubscribeRoom(roomId)
        super.onCleared()
    }
}
