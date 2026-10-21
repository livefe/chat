package com.nearby.chat.feature.support

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.encodeChatFile
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.SupportMessage
import com.nearby.chat.core.model.SupportSession
import com.nearby.chat.core.ui.Toaster
import com.nearby.chat.data.repo.ChatRepository
import com.nearby.chat.data.repo.SessionRepository
import com.nearby.chat.data.repo.SupportRepository
import com.nearby.chat.data.repo.UploadRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID
import javax.inject.Inject

@HiltViewModel
class SupportChatViewModel @Inject constructor(
    private val supportRepository: SupportRepository,
    private val uploadRepository: UploadRepository,
    private val chatRepository: ChatRepository,
    private val sessionRepository: SessionRepository,
    private val toaster: Toaster,
) : ViewModel() {
    private val _uiState = MutableStateFlow(
        SupportChatUiState(
            brandName = supportRepository.brandName,
            brandAvatar = supportRepository.brandAvatar,
        ),
    )
    val uiState: StateFlow<SupportChatUiState> = _uiState.asStateFlow()

    val displayMessages: StateFlow<List<MessageItem>> = _uiState
        .map { s -> s.messages.map { it.toChatMessage(s.selfId, s.brandAvatar) } }
        .stateIn(viewModelScope, SharingStarted.Eagerly, emptyList())

    init {
        viewModelScope.launch {
            supportRepository.incomingMessages.collect { m ->
                val sid = _uiState.value.session?.id ?: 0L
                if (sid > 0 && m.conversationId > 0 && m.conversationId != sid) return@collect
                appendLocal(m)
            }
        }
        viewModelScope.launch {
            supportRepository.sessionUpdates.collect { session ->
                _uiState.update {
                    it.copy(
                        session = session,
                        brandName = supportRepository.brandName,
                        brandAvatar = supportRepository.brandAvatar,
                    )
                }
            }
        }
        viewModelScope.launch {
            chatRepository.reconnectTick.collect {
                val sid = _uiState.value.session?.id ?: return@collect
                if (sid > 0) catchUp(sid)
            }
        }
    }

    fun enter() {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    loading = true,
                    selfId = sessionRepository.selfId(),
                    brandName = supportRepository.brandName,
                    brandAvatar = supportRepository.brandAvatar,
                )
            }
            supportRepository.enterViewing()
            runCatching { supportRepository.loadProfile() }
            val session = runCatching { supportRepository.refreshSession() }.getOrNull()
            val list = if (session != null) {
                runCatching { supportRepository.messages(session.id, limit = PAGE_LIMIT) }.getOrDefault(emptyList())
            } else {
                emptyList()
            }
            if (session != null) {
                runCatching { supportRepository.read(session.id, list.lastOrNull()?.id ?: 0L) }
            }
            _uiState.update {
                it.copy(
                    loading = false,
                    session = session,
                    messages = list,
                    brandName = supportRepository.brandName,
                    brandAvatar = supportRepository.brandAvatar,
                )
            }
        }
    }

    fun leave() {
        supportRepository.leaveViewing()
    }

    fun onInputChange(value: String) {
        _uiState.update { it.copy(input = value) }
    }

    fun send() {
        val text = _uiState.value.input.trim()
        if (text.isEmpty() || _uiState.value.sending || _uiState.value.uploading) return
        viewModelScope.launch {
            _uiState.update { it.copy(sending = true) }
            runCatching {
                val session = ensureSession()
                val msg = supportRepository.sendText(
                    conversationId = session.id,
                    text = text,
                    clientMsgId = newClientMsgId(),
                )
                supportRepository.noteLocalMessage(msg)
                appendLocal(msg)
                _uiState.update { it.copy(input = "", sending = false) }
            }.onFailure { e ->
                _uiState.update { it.copy(sending = false) }
                toaster.fail(e, R.string.chat_send_failed)
            }
        }
    }

    fun sendMedia(uri: Uri) {
        upload {
            val media = uploadRepository.uploadChatMedia(uri)
            dispatchMedia(media.contentType, media.mediaId.toString())
        }
    }

    fun sendFile(uri: Uri) {
        upload {
            val media = uploadRepository.uploadChatFile(uri)
            dispatchMedia(MessageContentTypes.FILE, encodeChatFile(media.mediaId, media.fileName))
        }
    }

    private fun upload(block: suspend () -> Unit) {
        if (_uiState.value.sending || _uiState.value.uploading) return
        viewModelScope.launch {
            _uiState.update { it.copy(uploading = true) }
            runCatching { block() }.onFailure { e -> toaster.fail(e, R.string.upload_failed) }
            _uiState.update { it.copy(uploading = false) }
        }
    }

    private suspend fun dispatchMedia(contentType: Int, content: String) {
        val session = ensureSession()
        val msg = supportRepository.sendMedia(
            conversationId = session.id,
            contentType = contentType,
            content = content,
            clientMsgId = newClientMsgId(),
        )
        supportRepository.noteLocalMessage(msg)
        appendLocal(msg)
    }

    private suspend fun ensureSession(): SupportSession {
        var session = _uiState.value.session
        if (session == null || session.status == SESSION_STATUS_CLOSED) {
            session = supportRepository.open(OPEN_SOURCE)
            val list = supportRepository.messages(session.id, limit = PAGE_LIMIT)
            _uiState.update { it.copy(session = session, messages = list) }
        }
        return session
    }

    private suspend fun catchUp(conversationId: Long) {
        runCatching {
            val afterId = _uiState.value.messages.maxOfOrNull { it.id } ?: 0L
            val list = supportRepository.messages(conversationId, beforeId = null, limit = PAGE_LIMIT)
            val newer = if (afterId <= 0L) list else list.filter { it.id > afterId }
            if (newer.isEmpty()) return
            _uiState.update { s ->
                if (s.session?.id != conversationId) return@update s
                s.copy(messages = mergeById(s.messages, newer))
            }
            val last = newer.lastOrNull()
            if (last != null) {
                supportRepository.noteLocalMessage(last)
                runCatching { supportRepository.read(conversationId, last.id) }
            }
        }
    }

    private fun mergeById(
        existing: List<SupportMessage>,
        incoming: List<SupportMessage>,
    ): List<SupportMessage> {
        if (incoming.isEmpty()) return existing
        if (existing.isEmpty()) return incoming
        val byId = LinkedHashMap<Long, SupportMessage>()
        for (m in existing) {
            if (m.id > 0) byId[m.id] = m
        }
        var changed = false
        for (m in incoming) {
            if (m.id <= 0) continue
            val prev = byId.put(m.id, m)
            if (prev == null || prev != m) changed = true
        }
        if (!changed) return existing
        return byId.values.sortedWith(compareBy({ it.createdAt }, { it.id }))
    }

    private fun appendLocal(m: SupportMessage) {
        _uiState.update { s ->
            if (m.id > 0 && s.messages.any { it.id == m.id }) s else s.copy(messages = s.messages + m)
        }
    }

    private fun newClientMsgId(): String = "u-${UUID.randomUUID()}"

    override fun onCleared() {
        leave()
        super.onCleared()
    }

    private companion object {
        const val PAGE_LIMIT = 50
        const val SESSION_STATUS_CLOSED = 2
        const val OPEN_SOURCE = "android"
    }
}
