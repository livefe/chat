package com.nearby.chat.data.repo

import android.content.Context
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.model.SupportMessage
import com.nearby.chat.core.model.SupportProfile
import com.nearby.chat.core.model.SupportSession
import com.nearby.chat.core.network.ensureSuccess
import com.nearby.chat.core.network.unwrap
import com.nearby.chat.data.mapper.toDomain
import com.nearby.chat.data.remote.api.SupportApi
import com.nearby.chat.data.remote.dto.SupportMessageDto
import com.nearby.chat.data.remote.dto.SupportOpenRequest
import com.nearby.chat.data.remote.dto.SupportReadRequest
import com.nearby.chat.data.remote.dto.SupportSendRequest
import com.nearby.chat.data.remote.dto.SupportSessionDto
import com.nearby.chat.data.remote.ws.WsScopes
import com.nearby.chat.data.remote.ws.WsTypes
import com.nearby.chat.data.remote.ws.decodeData
import com.nearby.chat.data.repo.chat.ChatConnection
import com.nearby.chat.di.ApplicationScope
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SupportRepository @Inject constructor(
    private val supportApi: SupportApi,
    private val connection: ChatConnection,
    @ApplicationScope private val scope: CoroutineScope,
    @ApplicationContext private val context: Context,
) {
    private val _profile = MutableStateFlow<SupportProfile?>(null)
    val profile: StateFlow<SupportProfile?> = _profile.asStateFlow()

    private val _session = MutableStateFlow<SupportSession?>(null)
    val session: StateFlow<SupportSession?> = _session.asStateFlow()

    private val _previewUnread = MutableStateFlow(0)
    val previewUnread: StateFlow<Int> = _previewUnread.asStateFlow()

    private val _incomingMessages = MutableSharedFlow<SupportMessage>(
        replay = 0,
        extraBufferCapacity = 64,
    )
    val incomingMessages: SharedFlow<SupportMessage> = _incomingMessages.asSharedFlow()

    private val _sessionUpdates = MutableSharedFlow<SupportSession>(extraBufferCapacity = 8)
    val sessionUpdates: SharedFlow<SupportSession> = _sessionUpdates.asSharedFlow()

    @Volatile
    private var viewing = false

    init {
        scope.launch {
            connection.messages.collect { msg ->
                when {
                    msg.type == WsTypes.SUPPORT_SESSION -> {
                        val session = msg.decodeData<SupportSessionDto>()?.toDomain() ?: return@collect
                        applySession(session, bumpUnread = !viewing)
                        _sessionUpdates.tryEmit(session)
                    }

                    msg.type == WsTypes.MESSAGE && msg.scope == WsScopes.SUPPORT -> {
                        val raw = msg.decodeData<SupportMessageDto>()?.toDomain() ?: return@collect
                        val m = raw.copy(
                            conversationId = raw.conversationId.takeIf { it > 0 }
                                ?: _session.value?.id
                                ?: 0L,
                        )
                        updatePreviewFromMessage(m)
                        if (viewing) {
                            _incomingMessages.tryEmit(m)
                            val sid = _session.value?.id ?: m.conversationId
                            if (sid > 0 && m.id > 0) {
                                runCatching { read(sid, m.id) }
                                _previewUnread.value = 0
                            }
                        } else {
                            _previewUnread.update { it + 1 }
                        }
                    }
                }
            }
        }
    }

    val brandName: String
        get() = _profile.value?.brandName?.takeIf { it.isNotBlank() }
            ?: _session.value?.brandName?.takeIf { it.isNotBlank() }
            ?: context.getString(R.string.support_brand_default)

    val brandAvatar: String
        get() = _profile.value?.brandAvatarUrl?.takeIf { it.isNotBlank() }
            ?: _session.value?.brandAvatarUrl.orEmpty()

    suspend fun loadProfile(): SupportProfile {
        val data = supportApi.profile().unwrap().toDomain()
        _profile.value = data
        return data
    }

    suspend fun refreshSession(): SupportSession? {
        val res = supportApi.current()
        res.ensureSuccess()
        val session = res.data?.toDomain()
        if (session != null) {
            applySession(session, bumpUnread = !viewing)
        } else {
            _session.value = null
        }
        return session
    }

    suspend fun open(source: String = "android"): SupportSession {
        val session = supportApi.open(SupportOpenRequest(source)).unwrap().toDomain()
        applySession(session, bumpUnread = false)
        return session
    }

    suspend fun messages(conversationId: Long, beforeId: Long? = null, limit: Int = 50): List<SupportMessage> = supportApi.messages(conversationId, beforeId, limit).unwrap().list.map { it.toDomain() }

    suspend fun sendText(conversationId: Long, text: String, clientMsgId: String): SupportMessage = sendMedia(conversationId, MessageContentTypes.TEXT, text, clientMsgId)

    suspend fun sendMedia(
        conversationId: Long,
        contentType: Int,
        content: String,
        clientMsgId: String,
    ): SupportMessage = supportApi.send(
        SupportSendRequest(
            conversationId = conversationId,
            contentType = contentType,
            content = content,
            clientMsgId = clientMsgId,
        ),
    ).unwrap().toDomain()

    suspend fun read(conversationId: Long, lastReadMsgId: Long = 0) {
        supportApi.read(SupportReadRequest(conversationId, lastReadMsgId)).ensureSuccess()
    }

    fun enterViewing() {
        viewing = true
        _previewUnread.value = 0
        connection.connect()
    }

    fun leaveViewing() {
        viewing = false
    }

    fun noteLocalMessage(m: SupportMessage) {
        updatePreviewFromMessage(m)
        val cur = _session.value
        if (cur != null) {
            _session.value = cur.copy(
                lastContent = previewOf(m),
                lastActiveAt = m.createdAt,
                lastMessageId = m.id,
            )
        }
    }

    private fun applySession(session: SupportSession, bumpUnread: Boolean) {
        _session.value = session
        if (bumpUnread) {
            _previewUnread.value = session.unreadCount.toInt().coerceAtLeast(0)
        }
    }

    private fun updatePreviewFromMessage(m: SupportMessage) {
        val cur = _session.value
        if (cur != null && (m.conversationId == 0L || m.conversationId == cur.id)) {
            _session.value = cur.copy(
                lastContent = previewOf(m),
                lastActiveAt = m.createdAt.takeIf { it > 0 } ?: cur.lastActiveAt,
                lastMessageId = m.id.takeIf { it > 0 } ?: cur.lastMessageId,
            )
        }
    }

    private fun previewOf(m: SupportMessage): String = if (m.recalled) {
        context.getString(R.string.msg_recalled_generic)
    } else {
        messagePreview(m.contentType, m.content).resolve(context)
    }
}
